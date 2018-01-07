package main

import (
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

const (
	downloadDir = "download"
	port        = ":50051"
)

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
	dbInfo.host = getEnvOrFlag("POSTGRES_HOST", "host", "127.0.0.1")
	dbInfo.port = getEnvOrFlag("POSTGRES_PORT", "port", "5432")
	dbInfo.usr = getEnvOrFlag("POSTGRES_USR", "usr", "postgres")
	dbInfo.pwd = getEnvOrFlag("POSTGRES_PWD", "pwd", "postgres")
	dbInfo.name = getEnvOrFlag("POSTGRES_DB", "db", "postgres")

	// health setup
	health.Register("internal errors", errorUpdater)
	health.RegisterPeriodicFunc("database_failure_check", time.Minute*2.5, func() error {
		err := dbPing()
		if err != nil {
			fmt.Println("database disconnected...")
			dbInfo.status = DB_ERR
		} else if dbInfo.status == DB_ERR {
			// there is no problem connecting with db,
			// but have internal error status
			// attempt to reconnect
			fmt.Println("database reconnecting...")
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

func getEnvOrFlag(key, flagstr, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return *flag.String(flagstr, fallback, flagstr)
	}
	return value
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
		fmt.Println("database connection failed, retry later...")
	} else {
		dbInfo.status = DB_UP
		fmt.Println("database connection successful")
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
