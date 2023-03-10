package service

import (
	"context"
	"errors"
	"log"
	"sync"

	"gocloudcamp_test/internal/database"
	"gocloudcamp_test/internal/playlist"
)

var (
	ErrNoPlaylistWithId = errors.New("there is no playlist with such id")
	ErrAlreadyExists    = errors.New("playlist with this id already exists")
)

type Playlists = map[uint]*playlist.Playlist

type Service struct {
	ChanForceStop chan struct{}
	db            *database.Database
	activeWg      sync.WaitGroup
	playlists     Playlists
	ChanError     chan error
}

func New(db *database.Database) *Service {
	service := &Service{}

	service.db = db

	service.ChanForceStop = make(chan struct{}, 1)
	service.ChanError = make(chan error)

	service.playlists = make(Playlists)

	return service
}

func (s *Service) Start() {
	log.Print("service | start")

	go func() {
		for err := range s.ChanError {
			if err != nil {
				log.Printf("service | error | %v", err)
			}
		}
	}()

	pls, err := s.db.LoadPlaylists()
	if err != nil {
		s.ChanError <- err

		return
	}

	sns, err := s.db.LoadSongs()
	if err != nil {
		s.ChanError <- err

		return
	}

	for _, pl := range pls {
		if err := s.AddPlaylist(pl.Id, pl.Name); err != nil {
			s.ChanError <- err

			continue
		}
	}

	for _, sn := range sns {
		if err := s.AddSong(sn.PlaylistId, sn.SongId, sn.Name, sn.Duration); err != nil {
			s.ChanError <- err

			continue
		}
	}
}

func (s *Service) ForceStop(cancel context.CancelFunc) {
	<-s.ChanForceStop

	log.Print("service | force stop")

	cancel()
}

func (s *Service) Stop(ctx context.Context) {
	<-ctx.Done()

	s.activeWg.Wait()

	log.Print("service | stop")
}

func (s *Service) GetPlaylists() Playlists {
	return s.playlists
}

func (s *Service) GetPlaylist(id uint) (*playlist.Playlist, error) {
	if pl, ok := s.playlists[id]; ok {
		return pl, nil
	}

	return nil, ErrNoPlaylistWithId
}

func (s *Service) LaunchPlaylist(ctx context.Context, id uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if pl.IsProcessing() {
		return playlist.ErrAlreadyProcessing
	}

	s.activeWg.Add(1)
	go func() {
		defer s.activeWg.Done()
		pl.Process(ctx)
	}()

	return nil
}

func (s *Service) CreatePlaylist(dbpl *database.Playlist) error {
	if err := s.db.CreatePlaylist(dbpl); err != nil {
		return err
	}

	return s.AddPlaylist(dbpl.Id, dbpl.Name)
}

func (s *Service) AddPlaylist(id uint, name string) error {
	if _, ok := s.playlists[id]; ok {
		return ErrAlreadyExists
	}

	pl := playlist.New(id, name)

	s.playlists[id] = pl

	return nil
}

func (s *Service) EditPlaylist(id uint, name string) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if err := s.db.UpdatePlaylist(id, name); err != nil {
		return err
	}

	pl.Name = name

	return nil
}

func (s *Service) DeletePlaylist(id uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if pl.IsProcessing() {
		if err := pl.Stop(); err != nil {
			return err
		}
	}

	if err := s.db.DeletePlaylist(id); err != nil {
		return err
	}

	for _, sn := range pl.GetSongsList() {
		if err := s.db.DeleteSong(sn.Id); err != nil {
			return err
		}
	}

	delete(s.playlists, id)

	return nil
}

func (s *Service) CreateSong(dbsn *database.Song) error {
	if err := s.db.CreateSong(dbsn); err != nil {
		return err
	}

	return s.AddSong(dbsn.PlaylistId, dbsn.SongId, dbsn.Name, dbsn.Duration)
}

func (s *Service) AddSong(id uint, sid uint, name string, duration uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	return pl.AddSong(sid, name, duration)
}

func (s *Service) EditSong(id uint, sid uint, name string, duration uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if pl.IsCurrent(sid) && pl.IsProcessing() {
		return playlist.ErrEditCurrent
	}

	sn, err := pl.GetSong(sid)
	if err != nil {
		return err
	}

	if name == "" {
		name = sn.Name
	}

	if duration == 0 {
		duration = sn.Duration
	}

	if err := s.db.UpdateSong(sid, name, duration); err != nil {
		return err
	}

	sn.Name = name
	sn.Duration = duration

	if pl.IsCurrent(sid) {
		return pl.SetTime(0)
	}

	return nil
}

func (s *Service) DeleteSong(id uint, sid uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if err := s.db.DeleteSong(sid); err != nil {
		return err
	}

	return pl.Remove(sid)
}
