package parsers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func FixMetadata(dbPath string, outPath string, p Parserer) error {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}
	posts := p.Scrap()
	if posts.GetLengthTable() == 0 {
		return errors.New("empty array post")
	}
	for i := 0; i < posts.GetLengthTable(); i++ {
		fileName := func() string {
			spl := strings.Split(posts.GetPostTable(i).FileUrl, "/")
			return strings.Split(spl[len(spl)-1], ".")[0] + ".json"
		}()
		f, err := os.OpenFile(
			filepath.Join(
				outPath,
				fileName,
			),
			os.O_RDWR|os.O_TRUNC, 0755,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer f.Close()
		mt := json.NewEncoder(f)
		err = mt.Encode(posts.GetPostTable(i))
		if err != nil {
			fmt.Println(err)
			continue
		}
		db.Model(&Post{}).Where("hash = ?", posts.GetPostTable(i).Hash).Update("tags", posts.GetPostTable(i).Tags)
		fmt.Printf("%s Fixed\n", fileName)
	}
	return nil
}
