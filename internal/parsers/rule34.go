package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Rulka struct {
	Url    string `json:"file_url"`
	Hash   string `json:"hash"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Rating string `json:"rating"`
	Score  int    `json:"score"`
	Tags   string `json:"tags"`
	Source string `json:"source"`
	Change int64  `json:"change"`
}

type Rule34Scraper struct {
	PostLimit    uint
	MaxPageLimit uint
	Tags         []string
	Proxy        string
	WaitTime     uint
	Logger       *log.Logger
}

func (r Rule34Scraper) convertPosts(p []Rulka) *PostTable {
	var posts *PostTable = NewPostTable()
	for i, j := range p {
		posts.AddPostTable(i, Post{
			Width:   j.Width,
			Height:  j.Height,
			Hash:    j.Hash,
			FileUrl: j.Url,
			Score:   j.Score,
			Rating:  string(j.Rating[0]),
			Tags:    j.Tags,
			FileExt: func() string {
				extSplit := strings.Split(j.Url, ".")
				return extSplit[len(extSplit)-1]
			}(),
			Sources:     strings.ReplaceAll(j.Source, " ", ","),
			DateCreated: time.Unix(j.Change, 0),
		})
	}
	return posts
}

func (r Rule34Scraper) Scrap() *PostTable {
	var (
		totPosts []Rulka
		tags     string
	)
	trs := http.Transport{
		Proxy: func() func(*http.Request) (*url.URL, error) {
			if r.Proxy != "" {
				proxy, err := url.Parse(r.Proxy)
				if err != nil {
					r.Logger.Printf("Cannot set proxy URL: %s", r.Proxy)
					r.Logger.Println(err)
					return nil
				}
				return http.ProxyURL(proxy)
			}
			return nil
		}(),
	}
	c := http.Client{
		Timeout:   10 * time.Second,
		Transport: &trs,
	}

	base_url := "https://api.rule34.xxx/index.php?page=dapi&s=post&q=index&json=1&"
	r.Logger.Println("Parsing tag list")
	for i, j := range r.Tags {
		j = strings.ReplaceAll(j, " ", "_")
		j = url.QueryEscape(j)
		tags += j
		if i < len(r.Tags)-1 {
			tags += "+"
		}
	}

	for i := 1; ; i++ {
		if r.MaxPageLimit > 0 && i > int(r.MaxPageLimit) {
			break
		}

		url := base_url + fmt.Sprintf("limit=%d&pid=%d&tags=%s",
			r.PostLimit, i-1, tags,
		)
		r.Logger.Printf("Visiting %s", url)
		resp, err := c.Get(url)
		if err != nil {
			r.Logger.Println(err)
			break
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var posts []Rulka
		err = json.Unmarshal(body, &posts)
		if err != nil {
			r.Logger.Println("Cannot parse response")
			r.Logger.Println(err)
			continue
		}
		if len(posts) == 0 {
			r.Logger.Println("Nothing found")
			break
		}

		r.Logger.Println("Appending posts to the list")
		totPosts = append(totPosts, posts...)
		time.Sleep(time.Duration(r.WaitTime) * time.Second)
	}
	return r.convertPosts(totPosts)
}
