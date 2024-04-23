package tagparser

import (
	"buster_daemon/e621PoolsDownloader/internal/proxy"
	"errors"
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

type DBTags struct {
	PostUrl string
	FileUrl string
	Tags    string
	FileExt string
	Rating  string
	Score   int
}

func (pt PostTags) ConvertToDB() *DBTags {
	dbView := DBTags{
		PostUrl: pt.PostUrl,
		FileUrl: pt.FileUrl,
		FileExt: pt.FileExt,
		Tags:    "",
		Rating:  pt.Rating,
		Score:   pt.Score,
	}

	for i, j := range pt.Tags {
		dbView.Tags += j
		if i < len(pt.Tags)-1 {
			dbView.Tags += ","
		}
	}

	return &dbView
}

func ParseTags(postUrl *PostTags, proxyUrl *url.URL, log *log.Logger) error {
	var (
		err      error
		splitUrl []string = strings.Split(postUrl.PostUrl, "?")
	)

	postUrl.PostUrl = splitUrl[0]
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
		splitUrl := strings.Split(h.Attr("href"), ".")
		postUrl.FileUrl = h.Attr("href")
		postUrl.FileExt = strings.ToLower(
			splitUrl[len(splitUrl)-1],
		)
		log.Println(
			"Adding file URL and it's extension: ",
			h.Attr("href"), splitUrl[len(splitUrl)-1])
	})

	coll.OnHTML("#post-rating-text", func(h *colly.HTMLElement) {
		postUrl.Rating = strings.ToLower(
			string(h.Text[0]),
		)
		log.Println("Adding rating: ", h.Text)
	})

	coll.Visit(postUrl.PostUrl)

	if postUrl.FileUrl == "" {
		log.Println("Post: ", postUrl.PostUrl)
		return errors.New("no file url, skipping")
	}

	return nil
}
