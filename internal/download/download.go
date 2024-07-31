package download

import (
	"bufio"
	"buster_daemon/boorus_downloader/internal/parsers"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Download struct {
	Proxy     string
	DBPath    string
	OutputDir string
	Wait      uint
	Logger    *log.Logger
}

type counters struct {
	downloaded uint
	skipped    uint
	failed     uint
}

func (d Download) DwPosts(p *parsers.PostTable) error {
	var (
		count counters = counters{
			downloaded: 0,
			skipped:    0,
			failed:     0,
		}
	)
	db, err := d.connectDB()
	if err != nil {
		log.Println(err)
		return err
	}
	d.ValidateAndExistence(p, db, &count)

	_, err = os.Stat(d.OutputDir)
	if err != nil {
		err = os.MkdirAll(d.OutputDir, 0755)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	trs := http.Transport{
		Proxy: func() func(*http.Request) (*url.URL, error) {
			if d.Proxy != "" {
				proxy, err := url.Parse(d.Proxy)
				if err == nil {
					return http.ProxyURL(proxy)
				}
				d.Logger.Panicln(err)
			}
			d.Logger.Println("Cannot set proxy")
			return nil
		}(),
	}
	c := http.Client{
		Transport: &trs,
		Timeout:   1200 * time.Second,
	}

	totProgress := progressbar.NewOptions(p.GetLengthTable(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription("Total Downloaded"),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWriter(os.Stderr),
	)

	for _, v := range p.GetMap() {
		// Initialise new buffer in RAM
		rbuf := new(bytes.Buffer)

		resp, err := c.Get(v.FileUrl)
		if err != nil {
			log.Println(err)
			totProgress.Add(1)
			count.skipped++
			count.failed++
			continue
		}

		resp.Request.Header.Set("User-Agent", "curl/8.8.0")
		defer resp.Body.Close()

		dwProgress := progressbar.DefaultBytes(resp.ContentLength,
			fmt.Sprintf(
				"Downloading: %s", v.Hash+"."+
					v.FileExt),
		)

		rspBuf := bufio.NewReader(resp.Body)
		var endErr = func(err error,
			progBar *progressbar.ProgressBar) {
			d.Logger.Println(err)
			progBar.Add(1)
			count.skipped++
			count.failed++
		}

		_, err = io.Copy(io.MultiWriter(
			rbuf, dwProgress,
		), rspBuf)
		if err != nil {
			endErr(err, totProgress)
			continue
		}

		_, err = rspBuf.WriteTo(rbuf)
		if err != nil {
			endErr(err, totProgress)
			continue
		}

		f, err := os.Create(filepath.Join(
			d.OutputDir, v.Hash+"."+
				v.FileExt,
		))
		if err != nil {
			d.Logger.Println(err)
			f.Close()
			count.skipped++
			continue
		}

		_, err = io.Copy(f, rbuf)
		if err != nil {
			d.Logger.Println(err)
			f.Close()
			os.Remove(f.Name())
			count.skipped++
			count.failed++
			continue
		}

		meta, err := os.Create(
			filepath.Join(
				d.OutputDir,
				func() string {
					sName := strings.Split(
						v.Hash, ".")
					return sName[0] + ".json"
				}(),
			),
		)
		if err != nil {
			d.Logger.Println(err)
			meta.Close()
			count.skipped++
			count.failed++
			continue
		}

		mt := json.NewEncoder(meta)
		mt.Encode(v)
		db.Create(v)
		totProgress.Add(1)
		f.Close()
		meta.Close()
		count.downloaded++
		time.Sleep(time.Duration(d.Wait) * time.Second)
	}

	fmt.Printf("\nDownloaded: %d files.\nSkipped: %d files.\n", count.downloaded,
		count.skipped)
	fmt.Printf("\n\nDownloaded: %d files.\nSkipped: %d files.\nFailed: %d files.", count.downloaded,
		count.skipped, count.failed)
	return nil
}

func (d Download) connectDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(d.DBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&parsers.Post{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d Download) ValidateAndExistence(postTable *parsers.PostTable,
	db *gorm.DB, c *counters) {
	var (
		postsCount int = postTable.GetLengthTable()
	)
	for i := 0; i < postsCount; i++ {
		if (parsers.Post{}) == postTable.GetPostTable(i) {
			fmt.Println("Invalid data retrieved, deleting from the list.")
			postTable.RemovePostTable(i)
			c.skipped++
			continue
		}
		if postTable.GetPostTable(i).FileUrl == "" {
			fmt.Println("File URL is empty, deleting from the list.")
			postTable.RemovePostTable(i)
			c.skipped++
			continue
		}
		res := db.Where("hash = ?", postTable.GetPostTable(i).Hash).
			First(&parsers.Post{
				Hash: postTable.GetPostTable(i).Hash,
			})
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			fmt.Println("Record exists, deleting from the list")
			postTable.RemovePostTable(i)
			c.skipped++
			continue
		}
	}
}
