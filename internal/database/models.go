package database

type Playlist struct {
	Id   uint   `json:",omitempty" gorm:"primarykey"`
	Name string `json:",omitempty" gorm:"default:playlist"`
}

type Song struct {
	SongId     uint   `json:",omitempty" gorm:"primarykey"`
	PlaylistId uint   `json:",omitempty"`
	Name       string `json:",omitempty" gorm:"default:song"`
	Duration   uint   `json:",omitempty" gorm:"default:1"`
}
