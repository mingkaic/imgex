//// file: imgex.go
// imgex downloads all images found on pages a certain depth from the input link's page
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/google/uuid"
	"github.com/mingkaic/imgdb"
	"github.com/mingkaic/phantomgo"
	"github.com/mingkaic/xcrawl"
	"gopkg.in/fatih/set.v0"
	"gopkg.in/yaml.v2"
)

type recordOptions struct {
	// first value = minimum image width,
	// second value = minimum image height,
	// ignores additional values
	// if second value is not available minimum height takes from first value
	limits []int `yaml:"image_limit"`
}

const downloadDir = "download"

var filenameGroup = regexp.MustCompile(`.+\/(.*)\.+`)
var db *imgdb.ImgDB

func main() {
	// argument sanitation
	flag.Parse()
	args := flag.Args()
	cyml := flag.String("cyml", "media_crawl.yml",
		"yml file outlining search constraint")
	if len(args) < 1 {
		log.Fatalf("Specify starting location")
	}

	// db setup
	db = imgdb.New("sqlite3", "test.db", downloadDir)
	defer db.Close()

	// search parameter components
	options, err := ioutil.ReadFile(*cyml)
	if err != nil {
		log.Fatalf("yaml read error: %v", err)
	}
	var rOpt recordOptions
	if err := yaml.Unmarshal(options, &rOpt); err == nil && len(rOpt.limits) > 0 {
		db.MinW = uint(rOpt.limits[0])
		if len(rOpt.limits) > 1 {
			db.MinW = uint(rOpt.limits[1])
		} else {
			db.MinH = db.MinW
		}
	}

	// crawl
	downloaded := set.New()
	crawler := xcrawl.New(options)
	crawler.InjectReq(xcrawl.GetDynamicRequest(phantomgo.GOPATH + "/src/github.com/mingkaic/imgex/bin/phantomjs"))
	crawler.InjectRec(func(wg *sync.WaitGroup, info *xcrawl.PageInfo) {
		defer wg.Done()
		baseLink, err := url.Parse(info.Link)
		if err != nil {
			log.Println(err)
			return
		}
		imgStews := info.DOM.FindAll("img")
		for _, img := range imgStews {
			if links, ok := img.Attrs["src"]; ok && len(links) > 0 {
				imgLink, err := url.Parse(links[0])
				if err != nil {
					log.Println(err)
					return
				}
				imgLink = baseLink.ResolveReference(imgLink)
				img := imgLink.String()
				if !downloaded.Has(img) {
					downloaded.Add(img)
					wg.Add(1) // wait on download
					go downloadLink(wg, img)
				}
			}
		}
	})
	crawler.Crawl(args[0])
	fmt.Println("image scraping complete")
}

func downloadLink(wg *sync.WaitGroup, link string) {
	defer wg.Done()
	resp, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	var b bytes.Buffer
	if _, err := io.Copy(&b, resp.Body); err != nil {
		log.Println(err)
		return
	}

	matches := filenameGroup.FindStringSubmatch(link)
	var name string
	if len(matches) > 1 {
		name = matches[1]
	} else {
		name = uuid.New().String()
	}
	_, err = db.AddImg(name, b)
	if err != nil {
		fmt.Println(link, err)
	} else {
		fmt.Println("successfully downloaded", link)
	}
}
