//// file: imgcrawl.go
// imgcrawl downloads all images found on pages a certain depth from the input link's page
// imgcrawl is used by server and console applications
package imgcrawl

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/google/uuid"
	"github.com/mingkaic/imgdb"
	"github.com/mingkaic/phantomgo"
	"github.com/mingkaic/xcrawl"
	"gopkg.in/fatih/set.v0"
)

// =============================================
//                    Declarations
// =============================================

type CrawlDB struct {
	*imgdb.ImgDB
}

type CrawlCB func(string, *imgdb.ImageFile, error)

// =============================================
//                    Globals
// =============================================

var filenameGroup = regexp.MustCompile(`.+\/(.*)\.+`)
var phantomPath = filepath.Join(phantomgo.GOPATH, "/src/github.com/mingkaic/imgex/bin/phantomjs")

// =============================================
//                    Public
// =============================================

// New ...
// Instantiates imgdb.ImgDB in CrawlDB
func New(dialect, source, filedir string) (out *CrawlDB, err error) {
	db, err := imgdb.New(dialect, source, filedir)
	if err != nil {
		return
	}
	out = &CrawlDB{db}
	return
}

// Crawl ...
// Core function
func (this CrawlDB) Crawl(link string, crawler *xcrawl.Crawler, cb CrawlCB) {
	context := set.New()
	crawler.InjectReq(xcrawl.GetDynamicRequest(phantomPath))
	crawler.InjectRec(func(wg *sync.WaitGroup, info *xcrawl.PageInfo) {
		defer wg.Done()
		baseLink, err := url.Parse(info.Link)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("processing", info.Link)
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
				if !context.Has(img) {
					context.Add(img)
					wg.Add(1) // wait on download
					go this.downloadLink(wg, img, cb)
				}
			}
		}
	})
	crawler.Crawl(link, context)
}

// =============================================
//                    Private
// =============================================

func (this CrawlDB) downloadLink(wg *sync.WaitGroup, link string, cb CrawlCB) {
	defer wg.Done()
	if this.SourceExists(link) {
		return
	}

	resp, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	matches := filenameGroup.FindStringSubmatch(link)
	var name string
	if len(matches) > 1 {
		name = matches[1]
	} else {
		name = uuid.New().String()
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	imgModel, err := this.AddImg(name, data)
	if err != nil {
		switch err.(type) {
		case *imgdb.DupFileError:
			this.AddSource(imgModel, link)
		}
	} else {
		this.AddSource(imgModel, link)
	}
	cb(link, imgModel, err)
}
