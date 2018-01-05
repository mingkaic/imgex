package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mingkaic/imgdb"
	"github.com/mingkaic/imgex/imgcrawl"
	"github.com/mingkaic/xcrawl"
)

const downloadDir = "download"

var GOPATH = os.Getenv("GOPATH")

func main() {
	flag.Parse()
	args := flag.Args()
	cyml := flag.String("cyml",
		filepath.Join(GOPATH, "src/github.com/mingkaic/imgex/console/media_crawl.yml"),
		"yml file outlining search constraint")
	if len(args) < 1 {
		log.Fatalf("Specify starting location")
	}

	// search parameter components
	options, err := ioutil.ReadFile(*cyml)
	if err != nil {
		log.Fatalf("yaml read error: %s", err)
	}

	// db setup
	db := imgcrawl.New("sqlite3", "imgex.db", downloadDir)
	defer db.Close()

	crawler := xcrawl.NewYaml(options)
	db.Crawl(args[0], crawler, func(link string, _ *imgdb.ImageFile, err error) {
		if err != nil {
			//fmt.Println(link, err)
		} else {
			fmt.Println("successfully downloaded", link)
		}
	})
	fmt.Println("image scraping complete")
}
