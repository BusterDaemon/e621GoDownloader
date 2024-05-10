package collector

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

type E621Collector struct {
	ProxyURL      *string
	PoolID        *int
	PostScrap     bool
	PostTags      *string
	MaxScrapPages *uint
	Logger        *log.Logger
	DB            *gorm.DB
}

func (c E621Collector) New(proxy *string, pool *int, postScrap bool,
	postTags *string, maxScrapPages *uint, logger *log.Logger,
	db *gorm.DB) Collectorer {
	return E621Collector{
		ProxyURL:      proxy,
		PoolID:        pool,
		PostScrap:     postScrap,
		PostTags:      postTags,
		MaxScrapPages: maxScrapPages,
		Logger:        logger,
		DB:            db,
	}
}

func (c E621Collector) GetProxy() string {
	return *c.ProxyURL
}

func (c E621Collector) GetPool() int {
	return *c.PoolID
}

func (c E621Collector) GetPostScrap() bool {
	return c.PostScrap
}

func (c E621Collector) GetPostTags() string {
	return *c.PostTags
}

func (c E621Collector) GetMaxScrapPages() uint {
	return *c.MaxScrapPages
}

func (c E621Collector) GetLogger() *log.Logger {
	return c.Logger
}

func (c E621Collector) GetDB() *gorm.DB {
	return c.DB
}

func (c E621Collector) SetProxy(url string) (Collectorer, error) {
	if strings.TrimSpace(url) == "" {
		return c, E621CollectorEmptyProxy{}
	}
	if !strings.Contains(url, "http") && !strings.Contains(url, "socks") {
		return c, E621CollectorUnknownProxy{}
	}
	c.ProxyURL = &url
	return c, nil
}

func (c E621Collector) SetPool(id int) (Collectorer, error) {
	if id == 0 || id < 0 {
		return c, E621CollectorZeroPoolID{}
	}
	c.PoolID = &id
	return c, nil
}

func (c E621Collector) SetPostTags(tags string) (Collectorer, error) {
	if strings.TrimSpace(tags) == "" {
		return c, E621CollectorEmptyTags{}
	}
	c.PostTags = &tags
	return c, nil
}

func (c E621Collector) SetLogger(logger *log.Logger) (Collectorer, error) {
	if logger == nil {
		return c, E621CollectorNullLogger{}
	}
	c.Logger = logger
	return c, nil
}

func (c E621Collector) SetDB(db *gorm.DB) (Collectorer, error) {
	if db == nil {
		return c, E621CollectorNullDB{}
	}
	c.DB = db
	return c, nil
}

func (c E621Collector) Scrap() ([]tagparser.PostTags, error) {
	var (
		pagesVisited uint = 0
		metas        []tagparser.PostTags
	)

	coll := colly.NewCollector(
		colly.AllowedDomains("e621.net"),
	)

	if c.ProxyURL != nil || *c.ProxyURL != "" {
		coll.SetProxy(*c.ProxyURL)
		c.Logger.Printf("Setted proxy: %s", *c.ProxyURL)
	}

	coll.OnHTML("article", func(h *colly.HTMLElement) {
		postUrl := h.Request.AbsoluteURL(
			"/posts/" + h.Attr("data-id"),
		)
		res := c.DB.Where("post_url = ?", postUrl).
			First(tagparser.DBTags{PostUrl: postUrl})
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) ||
			res.Error == nil {
			c.Logger.Println(postUrl, "Already downloaded. Skipping...")
			return
		}

		tags := strings.Split(h.Attr("data-tags"), " ")

		score, err := strconv.Atoi(h.Attr("data-score"))
		if err != nil {
			c.Logger.Println(err)
			return
		}
		metas = append(metas, tagparser.PostTags{
			PostUrl: postUrl,
			FileUrl: h.Attr("data-file-url"),
			Tags:    tags,
			FileExt: h.Attr("data-file-ext"),
			Rating:  h.Attr("data-rating"),
			Score:   score,
		})
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
			fmt.Sprintf("https://e621.net/%s", c.ParseTags()),
		)
		if err != nil {
			return nil, err
		}
	}
	return metas, nil
}

func (c E621Collector) ParseTags() string {
	var (
		result   string = "posts?tags="
		splitted []string
	)
	postTags := strings.ToLower(*c.PostTags)
	splitted = strings.Split(postTags, ",")
	for i, j := range splitted {
		j = strings.ReplaceAll(j, " ", "_")
		result += j
		if i != len(splitted)-1 {
			result += "+"
		}
	}
	result = strings.ReplaceAll(result, "\"", "")

	return result
}
