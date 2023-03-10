package database

import (
	"log"
)

func (db *Database) LoadPlaylists() ([]Playlist, error) {
	log.Print("database | load playlists")

	var pls []Playlist

	err := db.Order("id asc").Find(&pls).Error

	return pls, err
}

func (db *Database) CreatePlaylist(pl *Playlist) error {
	err := db.Create(&pl).Error

	log.Printf("database | create playlist | id %d", pl.Id)

	return err
}

func (db *Database) UpdatePlaylist(id uint, name string) error {
	pl := Playlist{Id: id}

	db.First(&pl)

	pl.Name = name

	log.Printf("database | update playlist | id %d", pl.Id)

	return db.Save(&pl).Error
}

func (db *Database) DeletePlaylist(id uint) error {
	log.Printf("database | delete playlist | id %d", id)

	return db.Delete(&Playlist{}, id).Error
}

func (db *Database) LoadSongs() ([]Song, error) {
	log.Print("database | load songs")

	var sns []Song

	err := db.Order("song_id asc").Find(&sns).Error

	return sns, err
}

func (db *Database) CreateSong(sn *Song) error {
	err := db.Create(&sn).Error

	log.Printf("database | create song | id %d", sn.SongId)

	return err
}

func (db *Database) UpdateSong(id uint, name string, duration uint) error {
	sn := Song{SongId: id}

	db.First(&sn)

	sn.Name = name
	sn.Duration = duration

	log.Printf("database | update song | id %d", sn.SongId)

	return db.Save(&sn).Error
}

func (db *Database) DeleteSong(id uint) error {
	log.Printf("database | delete song | id %d", id)

	return db.Delete(&Song{}, id).Error
}
