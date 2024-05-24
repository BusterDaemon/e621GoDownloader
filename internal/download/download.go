package download

import (
	"buster_daemon/e621PoolsDownloader/internal/parsers"
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
		Timeout:   30 * time.Second,
	}
	fNameUrl := func(url string) string {
		splUrl := strings.Split(url, "/")
		return splUrl[len(splUrl)-1]
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
			defer wg.Done()
			for j := start; j < end; j++ {
				res := db.Where("hash = ?", p[j].Hash).First(&parsers.Post{
					Hash: p[j].Hash,
				}).Order("file_url DESC")
				if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
					d.Logger.Println("Already downloaded skipping...")
					totProgress.Add(1)
					continue
				}

				f, err := os.CreateTemp(d.OutputDir, "*")
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
				defer resp.Body.Close()
				defer f.Close()

				dwProgress := progressbar.DefaultBytes(resp.ContentLength,
					fmt.Sprintf("Downloading: %s", fNameUrl(p[j].FileUrl)))

				_, err = io.Copy(io.MultiWriter(
					f, dwProgress,
				), resp.Body)
				if err != nil {
					log.Println(err)
					totProgress.Add(1)
					os.Remove(f.Name())
					continue
				}

				os.Rename(f.Name(), filepath.Join(
					d.OutputDir, fNameUrl(p[j].FileUrl),
				))
				meta, _ := os.Create(
					filepath.Join(
						d.OutputDir,
						func() string {
							sName := strings.Split(fNameUrl(p[j].FileUrl), ".")
							return sName[0] + ".json"
						}(),
					),
				)
				defer meta.Close()

				mt := json.NewEncoder(meta)
				mt.Encode(p[j])
				db.Create(p[j])
				totProgress.Add(1)
				if j < len(p)-1 {
					time.Sleep(time.Duration(d.Wait) * time.Second)
				}
			}
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
