package main

import (
	"buster_daemon/boorus_downloader/internal/download"
	"buster_daemon/boorus_downloader/internal/parsers"
	"fmt"
	"log"
	"os"
	"strings"

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
					&cli.BoolFlag{
						Name:  "dsort",
						Value: true,
						Usage: "use date sorting instead of score sorting (no effect for rule34)",
					},
					&cli.BoolFlag{
						Name:  "fix",
						Value: false,
						Usage: "fix metadata json files",
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
						dsort    bool     = ctx.Bool("dsort")
						fix      bool     = ctx.Bool("fix")
						posts    *parsers.PostTable
						parser   parsers.Parserer
					)
					if len(tags) == 1 {
						var tmpT []string
						t := strings.Split(tags[0], ";")
						tmpT = append(tmpT, t...)
						tags = tmpT
					}
					switch booru {
					case "e621":
						parser = parsers.E621Scraper{
							Proxy:        proxy,
							TagList:      tags,
							MaxPageLimit: maxPages,
							PostLimit:    maxPosts,
							WaitTime:     wait,
							DSorter:      dsort,
							Logger:       log.Default(),
						}
					case "rule34":
						parser = parsers.Rule34Scraper{
							PostLimit:    maxPosts,
							MaxPageLimit: maxPages,
							Tags:         tags,
							Proxy:        proxy,
							WaitTime:     wait,
							Logger:       log.Default(),
						}
					}
					switch fix {
					case true:
						err := parsers.FixMetadata(dbPath, outPath, parser)
						if err != nil {
							fmt.Println(err)
							return err
						}
					case false:
						posts = parser.Scrap()
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
					}
					return nil
				},
			},
		},
	}

	c.Run(os.Args)
}
