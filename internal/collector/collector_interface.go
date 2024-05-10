package collector

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"
	"log"

	"gorm.io/gorm"
)

type Collectorer interface {
	New(proxy *string, pool *int, postScrap bool,
		postTags *string, maxScrapPages *uint, logger *log.Logger,
		db *gorm.DB) Collectorer
	Scrap() ([]tagparser.PostTags, error)
	GetProxy() string
	GetPool() int
	GetPostScrap() bool
	GetPostTags() string
	GetMaxScrapPages() uint
	GetLogger() *log.Logger
	GetDB() *gorm.DB
	SetProxy(url string) (Collectorer, error)
	SetPool(id int) (Collectorer, error)
	SetPostTags(tags string) (Collectorer, error)
	SetLogger(logger *log.Logger) (Collectorer, error)
	SetDB(db *gorm.DB) (Collectorer, error)
}
