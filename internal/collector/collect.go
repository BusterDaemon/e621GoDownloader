package collector

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"buster_daemon/e621PoolsDownloader/internal/proxy"
	"fmt"
	"log"
	"net/url"

	"github.com/gocolly/colly"
)

func ScrapMetal(poolID int, proxyStr string, scrapPosts bool, postsTags string,
	maxPages uint, log *log.Logger) ([]tagparser.PostTags, error) {
	var (
		pagesVisited uint = 0
		urlPosts     []string
		metas        []tagparser.PostTags
	)

	proxyUrl, err := url.Parse(proxyStr)
	if err != nil {
		return nil, err
	}
	log.Println("Setted proxy: ", *proxyUrl)

	coll := colly.NewCollector(
		colly.AllowedDomains("e621.net"),
	)

	coll.WithTransport(
		proxy.DefaultTransport(proxyUrl),
	)

	coll.OnHTML("article > a[href]", func(h *colly.HTMLElement) {
		decodedUrl, _ := url.PathUnescape(
			h.Request.AbsoluteURL(h.Attr("href")),
		)
		if decodedUrl != "" {
			urlPosts = append(urlPosts, decodedUrl)
			log.Println("Added new post URL:", decodedUrl)
		}
	})

	coll.OnHTML("#paginator-next", func(h *colly.HTMLElement) {
		h.Request.Visit(h.Attr("href"))
		log.Println("Visiting", h.Request.AbsoluteURL(h.Attr("href")))
	})

	coll.OnRequest(func(r *colly.Request) {
		if maxPages > 0 && pagesVisited >= maxPages {
			r.Abort()
		}
		pagesVisited++
	})

	if !scrapPosts {
		log.Println("Scraping pool")
		coll.Visit(
			fmt.Sprintf("https://e621.net/pools/%d", poolID),
		)
	} else {
		log.Println("Scraping posts")
		coll.Visit(
			fmt.Sprintf("https://e621.net/%s", ParseTags(postsTags)),
		)
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
