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

type Danboorus struct {
	CreatedAt string `json:"created_at"`
	Source    string `json:"source"`
	Score     int    `json:"score"`
	Md5       string `json:"md5"`
	Rating    string `json:"rating"`
	Width     int    `json:"image_width"`
	Height    int    `json:"image_height"`
	Tags      string `json:"tag_string"`
	FileExt   string `json:"file_ext"`
	FileUrl   string `json:"file_url"`
}

type DanboorusScraper struct {
	BaseUrl      string
	PostLimit    uint
	MaxPageLimit uint
	Tags         []string
	Proxy        string
	WaitTime     uint
	Logger       *log.Logger
	ApiLogin     string
	ApiKey       string
}

func (r DanboorusScraper) convertPosts(p []Danboorus) *PostTable {
	var posts *PostTable = NewPostTable()
	for i, j := range p {
		createdDate, _ := time.Parse(time.RFC3339, j.CreatedAt)
		posts.AddPostTable(i, Post{
			Width:  j.Width,
			Height: j.Height,
			Hash:   j.Md5,
			FileUrl: func() string {
				if j.FileUrl != "" {
					return j.FileUrl
				}

				var baseUrl string = "https://cdn.donmai.us/original"

				return fmt.Sprintf("%s/%s/%c%c/%s.%s",
					baseUrl, j.Md5[:2], j.Md5[2], j.Md5[3], j.Md5, j.FileExt)
			}(),
			Score:       j.Score,
			Rating:      string(j.Rating[0]),
			Tags:        j.Tags,
			FileExt:     j.FileExt,
			Sources:     j.Source,
			DateCreated: createdDate,
		})
	}
	return posts
}

func (r DanboorusScraper) Scrap() *PostTable {
	var (
		totPosts []Danboorus
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

		url := r.BaseUrl + fmt.Sprintf("limit=%d&page=%d&tags=%s",
			r.PostLimit, i, tags,
		)
		r.Logger.Printf("Visiting %s", url)
		if r.ApiKey != "" {
			url = fmt.Sprintf("%s&login=%s&api_key=%s", url, r.ApiLogin, r.ApiKey)
		}

		resp, err := c.Get(url)
		if err != nil {
			r.Logger.Println(err)
			break
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var posts []Danboorus
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
