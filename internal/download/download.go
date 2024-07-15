package download

import (
	"bufio"
	"buster_daemon/boorus_downloader/internal/parsers"
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
	"sync"
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
	ParUnits  uint
	Logger    *log.Logger
}

func (d Download) DwPosts(p []parsers.Post) error {
	db, err := d.connectDB()
	if err != nil {
		log.Println(err)
		return err
	}

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

	totProgress := progressbar.NewOptions(len(p),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription("Total Downloaded"),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWriter(os.Stderr),
	)
	var wg sync.WaitGroup
	chunkSize := len(p) / int(d.ParUnits)

	for i := 0; i < int(d.ParUnits); i++ {
		wg.Add(1)
		go func(threadID int) {
			var start = i * chunkSize
			var end = start + chunkSize
			if threadID == int(d.ParUnits)-1 {
				end = len(p)
			}
			for j := start; j < end; j++ {
				if p[j].FileUrl == "" {
					d.Logger.Println("File URL is empty")
					totProgress.Add(1)
					continue
				}
				res := db.Where("hash = ?", p[j].Hash).First(&parsers.Post{
					Hash: p[j].Hash,
				}).Order("file_url DESC")
				if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
					d.Logger.Println("Already downloaded skipping...")
					totProgress.Add(1)
					continue
				}

				f, err := os.CreateTemp(d.OutputDir, "tmp-f")
				if err != nil {
					d.Logger.Println(err)
					totProgress.Add(1)
					continue
				}

				resp, err := c.Get(p[j].FileUrl)
				if err != nil {
					log.Println(err)
					totProgress.Add(1)
					continue
				}

				resp.Request.Header.Set("User-Agent", "curl/8.8.0")
				defer resp.Body.Close()
				defer f.Close()

				dwProgress := progressbar.DefaultBytes(resp.ContentLength,
					fmt.Sprintf("Downloading: %s", p[j].Hash+"."+p[j].FileExt))

				buf := bufio.NewReader(resp.Body)

				_, err = io.Copy(io.MultiWriter(
					f, dwProgress,
				), buf)
				var endErr = func(err error, file *os.File,
					progBar *progressbar.ProgressBar) {
					d.Logger.Println(err)
					progBar.Add(1)
					os.Remove(file.Name())
				}
				if err != nil {
					endErr(err, f, totProgress)
					continue
				}
				_, err = buf.WriteTo(f)
				if err != nil {
					endErr(err, f, totProgress)
					continue
				}

				os.Rename(f.Name(), filepath.Join(
					d.OutputDir, p[j].Hash+"."+p[j].FileExt,
				))
				meta, _ := os.Create(
					filepath.Join(
						d.OutputDir,
						func() string {
							sName := strings.Split(p[j].Hash, ".")
							return sName[0] + ".json"
						}(),
					),
				)
				defer meta.Close()

				mt := json.NewEncoder(meta)
				mt.Encode(p[j])
				db.Create(p[j])
				totProgress.Add(1)
				time.Sleep(time.Duration(d.Wait) * time.Second)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
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
