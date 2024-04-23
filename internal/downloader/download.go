package downloader

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"buster_daemon/e621PoolsDownloader/internal/proxy"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

func BatchDownload(urls []tagparser.PostTags, waitTime time.Duration, outDir string, proxyUrl string, log *log.Logger, scrapPosts *bool) error {
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err = os.MkdirAll(outDir, 0755); err != nil {
			return err
		}
	}
	parsProxy, _ := url.Parse(proxyUrl)
	cl := proxy.NewClient(proxyUrl)
	overallBar := progressbar.NewOptions(len(urls),
		progressbar.OptionSetDescription("Downloaded"),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWriter(os.Stderr))

	for i, j := range urls {
		tagparser.ParseTags(&j, parsProxy, log)

		resp, err := cl.Get(j.FileUrl)
		if err != nil {
			fmt.Printf("%s returned %s.\n", j.FileUrl, err)
			time.Sleep(waitTime * time.Second)
			continue
		}

		splitter := strings.Split(j.FileUrl, "/")

		var tmp_path string = path.Join(
			outDir,
			func() string {
				if *scrapPosts {
					return fmt.Sprintf("%s.tmp", splitter[len(splitter)-1])
				}
				return fmt.Sprintf("%.2d_%s.tmp", i, splitter[len(splitter)-1])
			}(),
		)
		var g_path string = strings.ReplaceAll(tmp_path, ".tmp", "")

		f, _ := os.OpenFile(
			tmp_path,
			os.O_CREATE|os.O_WRONLY,
			0644,
		)

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			fmt.Sprintf("Downloading: %s", j.FileUrl),
		)

		_, err = io.Copy(
			io.MultiWriter(f, bar),
			resp.Body,
		)
		if err != nil {
			os.Remove(tmp_path)
		}
		os.Rename(tmp_path, g_path)

		if err == nil {
			mt, _ := os.OpenFile(
				strings.ReplaceAll(
					g_path, path.Ext(g_path), ".json",
				),
				os.O_CREATE|os.O_WRONLY,
				0644,
			)
			js := json.NewEncoder(mt)
			js.Encode(j)
			defer mt.Close()
		}

		overallBar.Add(1)

		defer f.Close()

		if i != (len(urls) - 1) {
			time.Sleep(waitTime * time.Second)
		}
	}

	return nil
}
