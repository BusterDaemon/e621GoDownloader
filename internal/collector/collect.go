package collector

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

type Collector struct {
	ProxyURL      *string
	PoolID        *int
	PostScrap     bool
	PostTags      *string
	MaxScrapPages *uint
	Logger        *log.Logger
	DB            *gorm.DB
}

func (c Collector) Scrap() ([]tagparser.PostTags, error) {
	var (
		pagesVisited uint = 0
		urlPosts     []string
		metas        []tagparser.PostTags
	)

	coll := colly.NewCollector(
		colly.AllowedDomains("e621.net"),
	)

	if c.ProxyURL != nil || *c.ProxyURL != "" {
		coll.SetProxy(*c.ProxyURL)
		log.Println("Setted proxy: ", *c.ProxyURL)
	}

	coll.OnHTML("article > a[href]", func(h *colly.HTMLElement) {
		decodedUrl, _ := url.PathUnescape(
			h.Request.AbsoluteURL(h.Attr("href")),
		)

		if decodedUrl != "" {
			existing := c.DB.
				Where(
					"post_url = ?",
					strings.Split(decodedUrl, "?")[0],
				).
				First(
					&tagparser.DBTags{
						PostUrl: strings.Split(
							decodedUrl, "?",
						)[0],
					},
				)

			if !errors.Is(existing.Error, gorm.ErrRecordNotFound) ||
				existing.Error == nil {
				log.Printf("%s already downloaded, skipping...", decodedUrl)
				return
			}
			urlPosts = append(urlPosts, decodedUrl)
			log.Println("Added new post URL:", decodedUrl)
		}
	})

	coll.OnHTML("#paginator-next", func(h *colly.HTMLElement) {
		h.Request.Visit(h.Attr("href"))
		log.Println("Visiting", h.Request.AbsoluteURL(h.Attr("href")))
	})

	coll.OnRequest(func(r *colly.Request) {
		if *c.MaxScrapPages > 0 && pagesVisited >= *c.MaxScrapPages {
			r.Abort()
		}
		pagesVisited++
	})

	if !c.PostScrap || (c.PostTags == nil || *c.PostTags == "") {
		log.Println("Scraping pool")
		err := coll.Visit(
			fmt.Sprintf("https://e621.net/pools/%d", *c.PoolID),
		)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Scraping posts")
		err := coll.Visit(
			fmt.Sprintf("https://e621.net/%s", c.ParseTags(*c.PostTags)),
		)
		if err != nil {
			return nil, err
		}
	}

	log.Println("Adding posts URLs into metadata storage")
	for _, j := range urlPosts {
		if j != "" {
			metas = append(metas, tagparser.PostTags{
				PostUrl: j,
			})
		}
	}

	return metas, nil
}
