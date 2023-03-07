package database

import (
	"context"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

func Connect(ctx context.Context, uri string) *Database {
	log.Print("database | connecting")

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		log.Fatalf("database | %v", err)
	}

	db.AutoMigrate(&Playlist{}, &Song{})

	log.Print("database | connected")

	return &Database{db.WithContext(ctx)}
}
