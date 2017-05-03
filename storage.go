package main

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var storage *gorm.DB

func InitStorage() {
	path := "tmp"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	db, err := gorm.Open("sqlite3", "tmp/storage.db")
	if err != nil {
		log.Fatal("Cannot initialize storage: ", err.Error())
	}

	//db.LogMode(true)
	db.AutoMigrate(&Connection{}, &ActiveDatabase{})
	storage = db
}
