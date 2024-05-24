package main

import (
	"buster_daemon/boorus_downloader/internal/download"
	"buster_daemon/boorus_downloader/internal/parsers"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	var c = &cli.App{
		Name: "Booru Downloader",
		Commands: []*cli.Command{
			{
				Name:    "download",
				Aliases: []string{"dw"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "booru",
						Usage: "which booru to parse",
						Value: "e621",
					},
					&cli.StringSliceFlag{
						Name:     "tags",
						Required: true,
						Usage:    "tags for post searching",
					},
					&cli.StringFlag{
						Name:  "proxy",
						Usage: "proxy server to connect",
					},
					&cli.PathFlag{
						Name:  "db",
						Usage: "path to SQLite database",
						Value: "./downloaded.db",
					},
					&cli.PathFlag{
						Name:    "out",
						Aliases: []string{"o"},
						Usage:   "path to download directory",
						Value:   "./downloaded",
					},
					&cli.UintFlag{
						Name:  "posts",
						Value: 320,
						Usage: "maximum posts for page",
					},
					&cli.UintFlag{
						Name:  "pages",
						Value: 0,
						Usage: "maximum pages to parse (0 - unlimited)",
					},
					&cli.UintFlag{
						Name:  "wait",
						Value: 1,
						Usage: "wait time for downloads and API parsing",
					},
					&cli.UintFlag{
						Name:    "threads",
						Aliases: []string{"j"},
						Value:   1,
						Usage:   "how many threads will be download files",
					},
				},
				Action: func(ctx *cli.Context) error {
					var (
						booru    string   = ctx.String("booru")
						tags     []string = ctx.StringSlice("tags")
						proxy    string   = ctx.String("proxy")
						maxPosts uint     = ctx.Uint("posts")
						maxPages uint     = ctx.Uint("pages")
						wait     uint     = ctx.Uint("wait")
						dbPath   string   = ctx.Path("db")
						outPath  string   = ctx.Path("out")
						threads  uint     = ctx.Uint("threads")
						posts    []parsers.Post
					)
					switch booru {
					case "e621":
						posts = parsers.E621Scraper{
							Proxy:        proxy,
							TagList:      tags,
							MaxPageLimit: maxPages,
							PostLimit:    maxPosts,
							WaitTime:     wait,
							Logger:       log.Default(),
						}.Scrap()
					case "rule34":
						posts = parsers.Rule34Scraper{
							PostLimit:    maxPosts,
							MaxPageLimit: maxPages,
							Tags:         tags,
							Proxy:        proxy,
							WaitTime:     wait,
							Logger:       log.Default(),
						}.Scrap()
					}

					err := download.Download{
						Proxy:     proxy,
						DBPath:    dbPath,
						OutputDir: outPath,
						Wait:      wait,
						ParUnits:  threads,
						Logger:    log.Default(),
					}.DwPosts(posts)
					if err != nil {
						println(err)
						return err
					}
					return nil
				},
			},
		},
	}

	c.Run(os.Args)
}
