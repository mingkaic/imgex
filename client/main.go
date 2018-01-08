package main

import (
	"fmt"
	"io"
	"log"

	pb "github.com/mingkaic/imgex/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const address = "localhost:50051"

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCrawlerClient(conn)

	status, err := c.Health(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatal(err)
	}
	if status.ServiceUp {
		stream, err := c.Crawl(context.Background(), &pb.CrawlOpt{
			Link:     "http://www.reddit.com",
			MaxDepth: 12,
			SameHost: true,
		})
		if err != nil {
			log.Fatal(err)
		}
		for {
			imgInfo, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(imgInfo.ImgName)
		}
	} else {
		fmt.Println(status.Message)
	}
}
