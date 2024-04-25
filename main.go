package main

import (
	"buster_daemon/e621PoolsDownloader/internal/collector"
	"buster_daemon/e621PoolsDownloader/internal/database"
	"buster_daemon/e621PoolsDownloader/internal/downloader"
	"buster_daemon/e621PoolsDownloader/internal/env"
	"fmt"
	"log"
	"os"
	"path"

	flag "github.com/spf13/pflag"
)

func main() {
	poolID := flag.Int("poolID", 0, "Pool ID to download")
	waitTime := flag.Int("wait", 5, "Wait time between downloading (seconds)")
	scrapPosts := flag.Bool("scrapPosts", false, "Scrap posts or pools")
	postsTags := flag.String("pTags", "",
		"Search tags used to scrap posts. Delimited by commas.")
	maxPostPages := flag.Uint("maxPostPages", 0,
		"Maximum pages to scrap posts (0 = Unlimited)")
	outDir := flag.String("out", "./defOut/", "Output directory")
	proxyUrl := flag.String("proxy", "", "Proxy URL")
	dbPath := flag.String(
		"dbPath",
		path.Join(
			*outDir,
			"downloaded.db",
		),
		"Path to database that stores download history",
	)
	flag.Parse()

	logg := log.New(os.Stderr, "[DEBUG] ", 2)

	err := env.GetEnvData(waitTime, maxPostPages, outDir, proxyUrl, dbPath)
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

	db, err := database.New(*dbPath)
	if err != nil {
		panic(err)
	}

	urls, err := collector.Collector{
		ProxyURL:      proxyUrl,
		PoolID:        poolID,
		PostScrap:     *scrapPosts,
		PostTags:      postsTags,
		MaxScrapPages: maxPostPages,
		Logger:        logg,
		DB:            db,
	}.Scrap()
	if err != nil {
		log.Fatalln(err)
	}

	err = downloader.BatchDownload{
		Posts:            urls,
		WaitBtwDownloads: uint(*waitTime),
		OutputDir:        *outDir,
		ProxyUrl:         proxyUrl,
		Logger:           logg,
		ScrapPosts:       scrapPosts,
		DB:               db,
	}.Download()
	if err != nil {
		log.Fatalln(err)
	}
}
