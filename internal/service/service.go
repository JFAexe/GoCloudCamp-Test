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

type Service struct {
	activeWg      sync.WaitGroup
	DB            *database.Database
	ChanForceStop chan struct{}
	ChanError     chan error
	Playlists     map[uint]*playlist.Playlist
}

func New(db *database.Database) *Service {
	service := &Service{}

	service.DB = db

	service.ChanForceStop = make(chan struct{}, 1)
	service.ChanError = make(chan error)

	service.Playlists = make(map[uint]*playlist.Playlist)

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

	playlists, err := s.DB.LoadPlaylists()
	if err != nil {
		s.ChanError <- err

		return
	}

	songs, err := s.DB.LoadSongs()
	if err != nil {
		s.ChanError <- err

		return
	}

	for _, pl := range playlists {
		if err := s.AddPlaylist(pl.Id, pl.Name); err != nil {
			s.ChanError <- err

			continue
		}
	}

	for _, song := range songs {
		if err := s.AddSong(song.PlaylistId, song); err != nil {
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

func (s *Service) GetPlaylist(id uint) (*playlist.Playlist, error) {
	if pl, ok := s.Playlists[id]; ok {
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
	if err := s.DB.CreatePlaylist(dbpl); err != nil {
		return err
	}

	return s.AddPlaylist(dbpl.Id, dbpl.Name)
}

func (s *Service) AddPlaylist(id uint, name string) error {
	if _, ok := s.Playlists[id]; ok {
		return ErrAlreadyExists
	}

	pl := playlist.New(id, name)

	s.Playlists[id] = pl

	return nil
}

func (s *Service) EditPlaylist(id uint, name string) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if err := s.DB.UpdatePlaylist(id, name); err != nil {
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

	if err := s.DB.DeletePlaylist(id); err != nil {
		return err
	}

	for _, song := range pl.GetSongsList() {
		if err := s.DB.DeleteSong(song.Id); err != nil {
			return err
		}
	}

	delete(s.Playlists, id)

	return nil
}

func (s *Service) CreateSong(dbsn *database.Song) error {
	if err := s.DB.CreateSong(dbsn); err != nil {
		return err
	}

	return s.AddSong(dbsn.PlaylistId, *dbsn)
}

func (s *Service) AddSong(id uint, song database.Song) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	return pl.AddSong(song.SongId, song.Name, song.Duration)
}

func (s *Service) EditSong(id uint, sid uint, name string, duration uint) error {
	pl, err := s.GetPlaylist(id)
	if err != nil {
		return err
	}

	if pl.IsCurrent(sid) && pl.IsProcessing() {
		return playlist.ErrEditCurrent
	}

	song, err := pl.GetSong(sid)
	if err != nil {
		return err
	}

	if name == "" {
		name = song.Name
	}

	if duration == 0 {
		duration = song.Duration
	}

	if err := s.DB.UpdateSong(sid, name, duration); err != nil {
		return err
	}

	song.Name = name
	song.Duration = duration

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

	if err := s.DB.DeleteSong(sid); err != nil {
		return err
	}

	return pl.Remove(sid)
}
