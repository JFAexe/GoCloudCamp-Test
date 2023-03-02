package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

func Connect(uri string) *Database {
	log.Print("database | connecting")

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		log.Fatalf("database | %v", err)
	}

	db.AutoMigrate(&Playlist{}, &Song{})

	log.Print("database | connected")

	return &Database{db}
}
