package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"

	"github.com/mingkaic/imgdb"
	"github.com/mingkaic/imgex/imgcrawl"
	pb "github.com/mingkaic/imgex/proto"
	"github.com/mingkaic/xcrawl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct{}

const (
	downloadDir = "download"
	port        = ":50051"
)

var db *imgcrawl.CrawlDB
var filenameGroup = regexp.MustCompile(`.+\/(.*)\.+`)

func main() {
	db_host := getEnvOrFlag("POSTGRES_HOST", "host", "127.0.0.1")
	db_port := getEnvOrFlag("POSTGRES_PORT", "port", "5432")
	db_user := getEnvOrFlag("POSTGRES_USR", "usr", "postgres")
	db_password := getEnvOrFlag("POSTGRES_PWD", "pwd", "postgres")
	db_database := getEnvOrFlag("POSTGRES_DB", "db", "postgres")
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db_host, db_port, db_user, db_password, db_database)

	// db setup
	db = imgcrawl.New("postgres", psqlInfo, downloadDir)
	defer db.Close()

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

func (s *server) Crawl(in *pb.CrawlOpt, stream pb.Crawler_CrawlServer) error {
	crawler := &xcrawl.Crawler{
		MaxDepth:     uint(in.MaxDepth),
		SameHost:     in.SameHost,
		ContainsTags: in.ContainsTags,
	}
	db.Crawl(in.Link, crawler, func(link string, model *imgdb.ImageFile, err error) {
		if model != nil && err == nil {
			stream.Send(&pb.ImgInfo{
				ImgName: model.Name,
			})
		}
	})

	return nil
}

func getEnvOrFlag(key, flagstr, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return *flag.String(flagstr, fallback, flagstr)
	}
	return value
}
