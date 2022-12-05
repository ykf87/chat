package db

import (
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Create(filename string) (*gorm.DB, error) {
	_, err := os.Stat(filename)
	if err != nil {
		filename = strings.ReplaceAll(filename, "\\", "/")
		if strings.Contains(filename, "/") == true {
			rrs := strings.Split(filename, "/")
			rrs = rrs[:len(rrs)-1]
			dirs := strings.Join(rrs, "/")
			if _, err = os.Stat(dirs); err != nil {
				os.MkdirAll(dirs, 0755)
			}
		}
	}
	return gorm.Open(sqlite.Open(filename), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
}
