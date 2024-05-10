package collector

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

type Rule34Collector struct {
	ProxyURL      *string
	PoolID        *int
	PostScrap     bool
	PostTags      *string
	MaxScrapPages *uint
	Logger        *log.Logger
	DB            *gorm.DB
}

func (c *Rule34Collector) ParseTags() string {
	var (
		result string = `tags=`
	)

	tagsS := strings.Split(*c.PostTags, ",")
	for i, j := range tagsS {
		j = strings.ReplaceAll(j, " ", "_")
		result += j
		if i != len(tagsS)-1 {
			result += "+"
		}
	}
	result = strings.ReplaceAll(result, "\"", "")

	return result
}

func (c *Rule34Collector) getTheImage(url string) *tagparser.PostTags {
	coll := colly.NewCollector(
		colly.AllowedDomains("rule34.xxx"),
	)
	if *c.ProxyURL != "" && c.ProxyURL != nil {
		c.Logger.Printf("Setted proxy for post scraper: %s", *c.ProxyURL)
		coll.SetProxy(*c.ProxyURL)
	}

	var (
		post   tagparser.PostTags
		rating string
		score  int
		err    error
	)

	coll.OnHTML("#post-view", func(h *colly.HTMLElement) {
		var (
			imUrl string
			tags  []string
		)

		tagSb := h.DOM.Find("#tag-sidebar > .tag")
		tagSb.Each(func(_ int, s *goquery.Selection) {
			tagSel := s.Children().First().Next()
			tags = append(tags, strings.ReplaceAll(tagSel.Text(), " ", "_"))
		})

		links := h.DOM.Find(".link-list > ul > li > a")
		links.Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "Original image") {
				imUrl, _ = s.Attr("href")
				return
			}
		})

		imUrl = strings.Split(imUrl, "?")[0]
		c.Logger.Println("Found image source:", imUrl)

		stats := h.DOM.Find("#stats > ul > li")
		stats.Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "Rating") {
				rtSpl := strings.Split(strings.ToLower(s.Text()), " ")
				rating = string((rtSpl[1])[0])
				c.Logger.Println("Found image's rating:", rating)
				return
			}
			if strings.Contains(s.Text(), "Score") {
				ss := s.Children().First().Text()
				score, err = strconv.Atoi(ss)
				if err != nil {
					c.Logger.Println(err)
					return
				}
			}
		})

		c.Logger.Println("Found image's tags:", tags)

		imUrlSpl := strings.Split(imUrl, "/")
		extSpl := strings.Split(imUrlSpl[len(imUrlSpl)-1], ".")
		c.Logger.Println("Found image's extension:", extSpl[len(extSpl)-1])

		post = tagparser.PostTags{
			PostUrl: h.Request.URL.String(),
			FileUrl: imUrl,
			Tags:    tags,
			FileExt: extSpl[len(extSpl)-1],
			Rating:  rating,
			Score:   score,
		}
	})

	coll.OnRequest(func(r *colly.Request) {
		time.Sleep(time.Millisecond * 250)
	})

	coll.OnError(func(r *colly.Response, err error) {
	})

	err = coll.Visit(url)
	if err != nil {
		c.Logger.Println(err)
		return nil
	}

	if post.FileExt == "" ||
		post.FileUrl == "" ||
		post.PostUrl == "" ||
		post.Rating == "" ||
		len(post.Tags) == 0 {
		c.Logger.Println("Can't get the post data")
		return nil
	}
	return &post
}

func (c Rule34Collector) Scrap() ([]tagparser.PostTags, error) {
	var (
		pagesVisited uint = 0
		posts        []tagparser.PostTags
		visUrl       string
	)

	coll := colly.NewCollector(
		colly.AllowedDomains("rule34.xxx"),
	)

	if *c.ProxyURL != "" && c.ProxyURL != nil {
		c.Logger.Printf("Setted proxy: %s", *c.ProxyURL)
		coll.SetProxy(*c.ProxyURL)
	}

	coll.OnHTML(".thumb", func(h *colly.HTMLElement) {
		childLink := h.DOM.ChildrenFiltered("a")
		postUrl, ok := childLink.Attr("href")
		if !ok {
			c.Logger.Println("Can't get the post attrs")
			return
		}
		postUrl = h.Request.AbsoluteURL(postUrl)
		c.Logger.Printf("Found a post: %s", postUrl)
		post := c.getTheImage(postUrl)
		if post != nil {
			posts = append(posts, *post)
		}
	})

	coll.OnHTML(".pagination > a[href]", func(h *colly.HTMLElement) {
		if pagesVisited <= *c.MaxScrapPages {
			if h.Attr("alt") == "next" {
				pagesVisited++
				coll.Visit(h.Request.AbsoluteURL(h.Attr("href")))
			}
		}
	})

	coll.OnRequest(func(r *colly.Request) {
		c.Logger.Println("Visiting:", r.URL)
		if (c.MaxScrapPages != nil && *c.MaxScrapPages > 0) &&
			pagesVisited >= *c.MaxScrapPages {
			r.Abort()
		}
	})

	switch c.PostScrap {
	case true:
		tagsSpl := strings.Split(*c.PostTags, ",")
		for _, i := range tagsSpl {
			i = strings.ReplaceAll(i, " ", "_")
			visUrl += i
		}
		coll.Visit(
			fmt.Sprintf(`https://rule34.xxx/index.php?page=post&s=list&%s`, c.ParseTags()),
		)
	default:
		coll.Visit(
			fmt.Sprintf("https://rule34.xxx/index.php?page=pool&s=show&id=%d", *c.PoolID),
		)
	}

	return posts, nil
}
