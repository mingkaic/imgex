package main

import (
	"log"
	"net"
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
	// db setup
	db = imgcrawl.New("sqlite3", "test.db", downloadDir)
	defer db.Close()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCrawlerServer(s, &server{})
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
