package tagparser

import (
	"buster_daemon/e621PoolsDownloader/internal/proxy"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type PostTags struct {
	PostUrl string   `json:"postURL"`
	FileUrl string   `json:"fileURL"`
	Tags    []string `json:"tags"`
	FileExt string   `json:"fileExt"`
	Rating  string   `json:"rating"`
	Score   int      `json:"score"`
}

func ParseTags(postUrl *PostTags, proxyUrl *url.URL, log *log.Logger) {
	var err error

	coll := colly.NewCollector(
		colly.AllowedDomains("e621.net"),
	)
	coll.WithTransport(proxy.DefaultTransport(proxyUrl))

	coll.OnHTML(".search-tag", func(h *colly.HTMLElement) {
		postUrl.Tags = append(postUrl.Tags, h.Text)
		log.Println("Adding new tag: ", h.Text)
	})

	coll.OnHTML(".post-score", func(h *colly.HTMLElement) {
		postUrl.Score, err = strconv.Atoi(h.Text)
		if err != nil {
			return
		}
		log.Println("Adding score:", h.Text)
	})

	coll.OnHTML(".btn-warn", func(h *colly.HTMLElement) {
		decodedUrl, _ := url.PathUnescape(
			h.Request.AbsoluteURL(
				h.Attr("href"),
			),
		)
		splitUrl := strings.Split(decodedUrl, ".")
		postUrl.FileUrl = decodedUrl
		postUrl.FileExt = strings.ToLower(
			splitUrl[len(splitUrl)-1],
		)
		log.Println(
			"Adding file URL and it's extension: ",
			decodedUrl, splitUrl[len(splitUrl)-1])
	})

	coll.OnHTML("#post-rating-text", func(h *colly.HTMLElement) {
		postUrl.Rating = strings.ToLower(
			string(h.Text[0]),
		)
		log.Println("Adding rating: ", h.Text)
	})

	coll.Visit(postUrl.PostUrl)
}
