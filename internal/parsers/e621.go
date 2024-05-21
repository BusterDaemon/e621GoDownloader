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

type Eshka struct {
	// Id   int `json:"id"`
	File struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Ext    string `json:"ext"`
		// Size   int    `json:"size"`
		Hash string `json:"md5"`
		Url  string `json:"url"`
	} `json:"file"`
	Score struct {
		Total int `json:"total"`
	} `json:"score"`
	Tags struct {
		General   []string `json:"general"`
		Species   []string `json:"species"`
		Character []string `json:"character"`
		Artist    []string `json:"artist"`
		Lore      []string `json:"lore"`
		Meta      []string `json:"meta"`
	} `json:"tags"`
	Rating  string   `json:"rating"`
	Sources []string `json:"sources"`
	// Pools         []int    `json:"pools"`
	// Relationships struct {
	// 	ParentID    int   `json:"parent_id"`
	// 	HasChildren bool  `json:"has_children"`
	// 	Children    []int `json:"children"`
	// } `json:"relationships"`
}

type E621Posts struct {
	Post []Eshka `json:"posts"`
}

type E621Scraper struct {
	Proxy        string
	TagList      []string
	MaxPageLimit uint
	PostLimit    uint
	WaitTime     uint
	Logger       *log.Logger
}

func (s E621Posts) convert() []Post {
	var totalPosts []Post
	for _, i := range s.Post {
		totalPosts = append(totalPosts, Post{
			Width:  i.File.Width,
			Height: i.File.Height,
			Ext:    i.File.Ext,
			Hash:   i.File.Hash,
			Url:    i.File.Url,
			Score:  i.Score.Total,
			Rating: i.Rating,
			Tags: func() string {
				var massive []string
				massive = append(massive, i.Tags.Artist...)
				massive = append(massive, i.Tags.Character...)
				massive = append(massive, i.Tags.General...)
				massive = append(massive, i.Tags.Lore...)
				massive = append(massive, i.Tags.Meta...)
				massive = append(massive, i.Tags.Species...)
				return convertArray(massive)
			}(),
			Sources: convertArray(i.Sources),
		})
	}
	return totalPosts
}

func (s E621Scraper) Scrap() []Post {
	var (
		tagString string
		posts     E621Posts
	)

	trs := http.Transport{
		Proxy: func() func(*http.Request) (*url.URL, error) {
			if s.Proxy != "" {
				proxy, err := url.Parse(s.Proxy)
				if err != nil {
					s.Logger.Printf("Cannot set proxy URL: %s", s.Proxy)
					s.Logger.Println(err)
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

	s.Logger.Println("Parsing tag list", s.TagList)
	for j, k := range s.TagList {
		k = strings.ReplaceAll(k, " ", "_")
		tagString += k
		if j < len(s.TagList)-1 {
			tagString += "%20"
		}
	}

	for i := 1; ; i++ {
		if s.MaxPageLimit != 0 && i > int(s.MaxPageLimit) {
			break
		}
		url := fmt.Sprintf("https://e621.net/posts.json?limit=%d&page=%d&tags=%s%%20order:created_at",
			s.PostLimit, i, tagString,
		)

		s.Logger.Printf("Dialing: %s", url)
		resp, err := c.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var page_posts E621Posts
		err = json.Unmarshal(body, &page_posts)
		if err != nil {
			log.Println("Cannot parse response")
			log.Println(err)
			continue
		}

		if len(page_posts.Post) == 0 {
			break
		}

		s.Logger.Println("Appending posts to the list")
		posts.Post = append(posts.Post, page_posts.Post...)
		time.Sleep(time.Duration(s.WaitTime) * time.Second)
	}
	return posts.convert()
}
