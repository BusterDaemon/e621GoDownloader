package database

import (
	tagparser "buster_daemon/e621PoolsDownloader/internal/collector/tag_parser"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(
		sqlite.Open(dbPath),
		&gorm.Config{},
	)
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&tagparser.DBTags{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
