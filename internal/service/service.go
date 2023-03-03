package service

import (
	"context"
	"errors"
	"log"

	"gocloudcamp_test/internal/database"
	"gocloudcamp_test/internal/playlist"
)

type Service struct {
	DB        *database.Database
	Errs      chan error
	Playlists map[uint]*playlist.Playlist
	Songs     map[uint]*playlist.Song
}

func New(db *database.Database) *Service {
	service := &Service{}

	service.DB = db
	service.Errs = make(chan error)

	service.Playlists = make(map[uint]*playlist.Playlist)
	service.Songs = make(map[uint]*playlist.Song)

	return service
}

func (s *Service) Start() {
	log.Print("service | start")

	go func() {
		for err := range s.Errs {
			if err != nil {
				log.Printf("service | error | %v", err)
			}
		}
	}()

	pls, err := s.DB.LoadPlaylists()
	if err != nil {
		s.Errs <- err

		return
	}

	sns, err := s.DB.LoadSongs()
	if err != nil {
		s.Errs <- err

		return
	}

	for _, pl := range pls {
		err := s.AddPlaylist(pl.Id, pl.Name)
		if err != nil {
			s.Errs <- err

			continue
		}
	}

	for _, song := range sns {
		s.AddSong(song.PlaylistId, song)
	}
}

func (s *Service) Stop(ctx context.Context) {
	<-ctx.Done()

	for _, pl := range s.Playlists {
		s.DB.UpdatePlaylist(&database.Playlist{Id: pl.Id, Name: pl.Name})
	}

	log.Print("service | stop")
}

func (s *Service) GetPlaylist(id uint) (*playlist.Playlist, error) {
	if pl, ok := s.Playlists[id]; ok {
		return pl, nil
	}

	return nil, errors.New("there is no playlist with such id")
}

func (s *Service) LaunchPlaylist(ctx context.Context, id uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if pl.Status().Processing {
		return errors.New("can't launch already processing playlist")
	}

	go pl.Process(ctx)

	return nil
}

func (s *Service) AddPlaylist(id uint, name string) error {
	if _, ok := s.Playlists[id]; ok {
		return errors.New("can't create playlist with existing id")
	}

	pl := playlist.New(id, name)

	s.Playlists[id] = pl

	return nil
}

func (s *Service) EditPlaylist(id uint, name string) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		s.Errs <- err

		return err
	}

	pl.Name = name

	return nil
}

func (s *Service) DeletePlaylist(id uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		s.Errs <- err

		return err
	}

	if err := pl.Stop(); err != nil {
		s.Errs <- err

		return err
	}

	delete(s.Playlists, id)

	return nil
}

func (s *Service) AddSong(id uint, song database.Song) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		s.Errs <- err

		return err
	}

	sn, err := pl.AddSong(song.SongId, song.Name, song.Duration)
	if err != nil {
		s.Errs <- err

		return err
	}

	s.Songs[song.SongId] = sn

	return nil
}

func (s *Service) EditSong(id uint, sid uint, song database.Song) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		s.Errs <- err

		return err
	}

	if pl.Status().CurrentId == sid {
		return errors.New("can't edit playing song")
	}

	sn, err := pl.GetSong(sid)
	if err != nil {
		s.Errs <- err

		return err
	}

	sn.Name = song.Name
	sn.Duration = song.Duration

	return nil
}

func (s *Service) DeleteSong(id uint, sid uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		s.Errs <- err

		return err
	}

	err = pl.Remove(sid)
	if err != nil {
		s.Errs <- err

		return err
	}

	delete(s.Songs, sid)

	return nil
}
