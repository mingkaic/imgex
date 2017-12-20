//// file: imgex.go
// imgex downloads all images found on pages a certain depth from the input link's page
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/google/uuid"
	"github.com/mingkaic/xcrawl"
	"gopkg.in/fatih/set.v0"
	"gopkg.in/h2non/filetype.v1"
	"gopkg.in/yaml.v2"
)

const (
	DEFAULT_WIDTH = 500
	DEFAULT_HEIGHT = 500
)

var filenameGroup = regexp.MustCompile(`.+\/(.*)\.+`)

var IMG_W = DEFAULT_WIDTH
var IMG_H = DEFAULT_HEIGHT

type recordOptions struct {
	// first value = minimum image width,
	// second value = minimum image height,
	// ignores additional values
	// if second value is not available minimum height takes from first value
	limits []int `yaml:"image_limit"`
}

func main() {
	// argument sanitation
	flag.Parse()
	args := flag.Args()
	cyml := flag.String("cyml", "media_crawl.yml",
		"yml file outlining search constraint")
	if len(args) < 1 {
		log.Fatalf("Specify starting location")
	}

	// search parameter components
	options, err := ioutil.ReadFile(*cyml)
	if err != nil {
		log.Fatalf("yaml read error: %v", err)
	}
	var rOpt recordOptions
	if err := yaml.Unmarshal(options, &rOpt); err == nil && len(rOpt.limits) > 0 {
		IMG_W = rOpt.limits[0]
		if len(rOpt.limits) > 1 {
			IMG_H = rOpt.limits[1]
		} else {
			IMG_H = IMG_W
		}
	}

	downloaded := set.New()
	downloadDir := "download"
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		err = os.MkdirAll(downloadDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	// crawl
	crawler := xcrawl.New(options)
	crawler.Record(func(info *xcrawl.PageInfo) {
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
					go downloadLink(img, downloadDir)
				}
			}
		}
	})
	crawler.Crawl(args[0])
}

func downloadLink(link, downloadDir string) {
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
	rawdata := b.Bytes()

	kind, err := filetype.Match(rawdata)
	if err != nil {
		log.Println("Unknown file: %s", err)
		return
	}

	img, _, err := image.Decode(&b)
	tooSmall := false
	if err == nil {
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		tooSmall = width < IMG_W || height < IMG_H
	}

	if tooSmall {
		return
	}
	matches := filenameGroup.FindStringSubmatch(link)
	var name string
	if len(matches) > 1 {
		name = matches[1]
		if _, err := os.Stat(downloadDir + "/" + name + "." + kind.Extension); err == nil {
			name += uuid.New().String()
		}
	} else {
		name = uuid.New().String()
	}
	filepath := downloadDir + "/" + name + "." + kind.Extension
	file, err := os.Create(filepath)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	_, err = file.Write(rawdata)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(filepath, " saved")
}
