package downloader

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"buster_daemon/e621PoolsDownloader/internal/proxy"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
)

type BatchDownload struct {
	Posts            []tagparser.PostTags
	WaitBtwDownloads uint
	OutputDir        string
	ProxyUrl         *string
	Logger           *log.Logger
	ScrapPosts       *bool
	DB               *gorm.DB
}

type Downloader interface {
	Download() error
}

func (bd BatchDownload) Error() string {
	return "Post list is empty.\n"
}

func (bd BatchDownload) Download() error {
	if len(bd.Posts) == 0 {
		return &BatchDownload{}
	}

	if _, err := os.Stat(bd.OutputDir); os.IsNotExist(err) {
		if err = os.MkdirAll(bd.OutputDir, 0755); err != nil {
			return err
		}
	}

	cl := proxy.NewClient(*bd.ProxyUrl)
	overallBar := progressbar.NewOptions(len(bd.Posts),
		progressbar.OptionSetDescription("Downloaded"),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWriter(os.Stderr))

	for i, j := range bd.Posts {
		postUrl := strings.Split(j.GetPostUrl(), "?")
		j.SetPostUrl(postUrl[0])

		srch := tagparser.DBTags{PostUrl: j.GetPostUrl()}
		res := bd.DB.Where("post_url = ?", j.GetPostUrl()).First(&srch)
		if res.Error == nil {
			log.Println("Already downloaded, skipping...")
			overallBar.Add(1)
			continue
		}

		err := j.ParseTags(bd.ProxyUrl, bd.Logger)
		if err != nil {
			log.Println(err)
			overallBar.Add(1)
			continue
		}

		splitter := strings.Split(j.GetFileUrl(), "/")

		resp, err := cl.Get(j.GetFileUrl())
		if err != nil {
			fmt.Printf("%s returned %s.\n", j.GetFileUrl(), err)
			time.Sleep(time.Duration(bd.WaitBtwDownloads) * time.Second)
			continue
		}

		var tmp_path string = path.Join(
			bd.OutputDir,
			func() string {
				if bd.ScrapPosts != nil && *bd.ScrapPosts {
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
			fmt.Sprintf("Downloading: %s", j.GetFileUrl()),
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
		bd.DB.Create(j.ConvertToDB())

		defer f.Close()

		if i != (len(bd.Posts) - 1) {
			time.Sleep(time.Duration(bd.WaitBtwDownloads) * time.Second)
		}
	}

	return nil
}
