package downloader

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"buster_daemon/e621PoolsDownloader/internal/proxy"
	"crypto/sha512"
	"encoding/json"
	"errors"
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
	WaitBtwDownloads uint
	OutputDir        string
	ProxyUrl         *string
	Logger           *log.Logger
	ScrapPosts       *bool
	DB               *gorm.DB
}

func (bd BatchDownload) Error() string {
	return "Post list is empty.\n"
}

func (bd BatchDownload) Download(posts []tagparser.PostTags) error {
	if len(posts) == 0 {
		return &BatchDownload{}
	}

	if _, err := os.Stat(bd.OutputDir); os.IsNotExist(err) {
		if err = os.MkdirAll(bd.OutputDir, 0755); err != nil {
			return err
		}
	}

	cl := proxy.NewClient(*bd.ProxyUrl)
	overallBar := progressbar.NewOptions(len(posts),
		progressbar.OptionSetDescription("Downloaded"),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWriter(os.Stderr))

	for i, j := range posts {
		srch := tagparser.DBTags{PostUrl: j.PostUrl}
		res := bd.DB.Where("post_url = ?", j.PostUrl).First(&srch)
		if res.Error == nil {
			log.Println("Already downloaded, skipping...")
			overallBar.Add(1)
			continue
		}

		splitter := strings.Split(j.FileUrl, "/")

		resp, err := cl.Get(j.FileUrl)
		if err != nil {
			fmt.Printf("%s returned %s.\n", j.FileUrl, err)
			time.Sleep(time.Duration(bd.WaitBtwDownloads) * time.Second)
			continue
		}

		tmpFile, err := os.CreateTemp(bd.OutputDir, "*")
		if err != nil {
			return err
		}
		defer tmpFile.Close()
		tmpFilePath := tmpFile.Name()

		newFileName := func() string {
			if bd.ScrapPosts != nil && *bd.ScrapPosts {
				return path.Join(
					bd.OutputDir,
					splitter[len(splitter)-1],
				)
			}
			return path.Join(
				bd.OutputDir,
				fmt.Sprintf("%.3d_%s", i, splitter[len(splitter)-1]),
			)
		}

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			fmt.Sprintf("Downloading: %s", j.FileUrl),
		)

		_, err = io.Copy(
			io.MultiWriter(tmpFile, bar),
			resp.Body,
		)
		if err != nil {
			os.Remove(tmpFilePath)
		}
		_, err = tmpFile.Seek(0, 0)
		if err != nil {
			return err
		}

		h := sha512.New()
		_, err = io.Copy(h, tmpFile)
		if err != nil {
			return err
		}
		hstring := fmt.Sprintf("%x", h.Sum(nil))

		res = bd.DB.Where("hash = ?",
			hstring,
		).First(&tagparser.DBTags{
			Hash: hstring,
		})
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) ||
			res.Error == nil {
			os.Remove(tmpFilePath)
			overallBar.Add(1)
			bd.DB.Create(j.ConvertToDB(fmt.Sprintf("%x", h.Sum(nil))))
			continue
		}

		err = os.Rename(tmpFilePath, newFileName())
		if err == nil {
			mt, _ := os.OpenFile(
				strings.ReplaceAll(
					newFileName(), path.Ext(newFileName()), ".json",
				),
				os.O_CREATE|os.O_WRONLY,
				0644,
			)
			js := json.NewEncoder(mt)
			js.Encode(j)
			defer mt.Close()
		}

		overallBar.Add(1)
		bd.DB.Create(j.ConvertToDB(fmt.Sprintf("%x", h.Sum(nil))))

		if i != (len(posts) - 1) {
			time.Sleep(time.Duration(bd.WaitBtwDownloads) * time.Second)
		}
	}

	return nil
}
