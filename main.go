package main

import (
	"buster_daemon/e621PoolsDownloader/internal/collector"
	"buster_daemon/e621PoolsDownloader/internal/downloader"
	"buster_daemon/e621PoolsDownloader/internal/env"
	"fmt"
	"log"
	"os"
	"time"

	flag "github.com/spf13/pflag"
)

func main() {
	poolID := flag.Int("poolID", 0, "Pool ID to download")
	waitTime := flag.Int("wait", 5, "Wait time between downloading (seconds)")
	scrapPosts := flag.Bool("scrapPosts", false, "Scrap posts or pools")
	postsTags := flag.String("pTags", "", "Search tags used to scrap posts. Delimited by commas.")
	maxPostPages := flag.Uint("maxPostPages", 0, "Maximum pages to scrap posts (0 = Unlimited)")
	outDir := flag.String("out", "./defOut/", "Output directory")
	proxyUrl := flag.String("proxy", "", "Proxy URL")
	flag.Parse()

	logg := log.New(os.Stderr, "[DEBUG]", 2)

	err := env.GetEnvData(waitTime, maxPostPages, outDir, proxyUrl)
	if err != nil {
		fmt.Println(err)
	}

	if *poolID == 0 && !*scrapPosts {
		flag.Usage()
		return
	}

	if *postsTags == "" && *scrapPosts {
		flag.Usage()
		return
	}

	urls, err := collector.ScrapMetal(*poolID, *proxyUrl, *scrapPosts, *postsTags, *maxPostPages, logg)
	if err != nil {
		panic(err)
	}

	err = downloader.BatchDownload(urls, time.Duration(*waitTime), *outDir, *proxyUrl, logg, scrapPosts)
	if err != nil {
		panic(err)
	}
}
