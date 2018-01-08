package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/docker/go-healthcheck"
	"github.com/mingkaic/imgdb"
	"github.com/mingkaic/imgex/imgcrawl"
	pb "github.com/mingkaic/imgex/proto"
	"github.com/mingkaic/xcrawl"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// =============================================
//                    Declarations
// =============================================

type server struct{}

const (
	DB_DOWN = iota // database is down (expected)
	DB_UP   = iota // database is up
	DB_ERR  = iota // database is down (unexpected)
)

// =============================================
//                    Globals
// =============================================

const port = ":50051"

var downloadDir string
var filenameGroup = regexp.MustCompile(`.+\/(.*)\.+`)
var errorUpdater = health.NewStatusUpdater()
var dbInfo = struct {
	db             *imgcrawl.CrawlDB
	host, port     string
	usr, pwd, name string
	status         int
}{status: DB_DOWN}

// =============================================
//                    Server
// =============================================

func main() {
	flag.StringVar(&downloadDir, "download", "./download",
		"directory path to copy download images")
	getEnvOrFlag(&dbInfo.host, "POSTGRES_HOST", "host", "127.0.0.1")
	getEnvOrFlag(&dbInfo.port, "POSTGRES_PORT", "port", "5432")
	getEnvOrFlag(&dbInfo.usr, "POSTGRES_USR", "usr", "postgres")
	getEnvOrFlag(&dbInfo.pwd, "POSTGRES_PWD", "pwd", "postgres")
	getEnvOrFlag(&dbInfo.name, "POSTGRES_DB", "db", "postgres")
	flag.Parse()

	// health setup
	health.Register("internal errors", errorUpdater)
	health.RegisterPeriodicFunc("database_failure_check", time.Minute*3, func() error {
		err := dbPing()
		if err != nil {
			log.Println("database disconnected...")
			dbInfo.status = DB_ERR
		} else if dbInfo.status == DB_ERR {
			// there is no problem connecting with db,
			// but have internal error status
			// attempt to reconnect
			log.Println("database reconnecting...")
			connDB()
		}
		return err
	})

	// db setup
	connDB()
	defer closeDB()

	// grpc setup
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCrawlerServer(s, &server{})
	fmt.Printf("server listening on %s\n", port)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Argument Utility

func getEnvOrFlag(out *string, key, flagStr, fallback string) {
	value := os.Getenv(key)
	if len(value) == 0 {
		flag.StringVar(out, flagStr, fallback, flagStr)
	} else {
		*out = value
	}
}

// Database utility

func connDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbInfo.host, dbInfo.port, dbInfo.usr, dbInfo.pwd, dbInfo.name)
	var err error
	dbInfo.db, err = imgcrawl.New("postgres", psqlInfo, downloadDir)
	if err != nil {
		log.Println(err)
		errorUpdater.Update(err)
		dbInfo.status = DB_ERR
		log.Println("database connection failed, retry later...")
	} else {
		dbInfo.status = DB_UP
		log.Println("database connection successful")
	}
}

func dbPing() error {
	addr := dbInfo.host + ":" + dbInfo.port
	conn, err := net.DialTimeout("tcp", addr, time.Minute)
	if err != nil {
		return errors.New("connection to " + addr + " failed")
	}
	conn.Close()
	return err
}

func closeDB() {
	if dbInfo.db != nil {
		dbInfo.db.Close()
	} else {
		dbInfo.status = DB_DOWN
	}
}

// =============================================
//                    GRPC API
// =============================================

func (s *server) Crawl(in *pb.CrawlOpt, stream pb.Crawler_CrawlServer) error {
	crawler := &xcrawl.Crawler{
		MaxDepth:     uint(in.MaxDepth),
		SameHost:     in.SameHost,
		ContainsTags: in.ContainsTags,
	}
	dbInfo.db.Crawl(in.Link, crawler, func(link string, model *imgdb.ImageFile, err error) {
		if model != nil && err == nil {
			stream.Send(&pb.ImgInfo{
				ImgName: model.Name,
			})
		}
	})

	return nil
}

func (s *server) Health(ctx context.Context, _ *pb.Empty) (*pb.ServiceStatus, error) {
	stats := health.CheckStatus()
	jstr, err := json.Marshal(stats)
	return &pb.ServiceStatus{ServiceUp: len(stats) == 0, Message: string(jstr)}, err
}
