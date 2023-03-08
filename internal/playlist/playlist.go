package playlist

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrNotProcessed      = errors.New("playlist is not being processed")
	ErrAlreadyProcessing = errors.New("playlist is already processing")
	ErrAlreadyStopped    = errors.New("playlist is already stopped")
	ErrAlreadyPlaying    = errors.New("playlist is already playing")
	ErrAlreadyPaused     = errors.New("playlist is already paused")
	ErrSwitchLast        = errors.New("this is the last song")
	ErrSwitchFirst       = errors.New("this is the first song")
	ErrSongNotIn         = errors.New("song is not in playlist")
	ErrSongIdTaken       = errors.New("song with this id is already in playlist")
	ErrRemoveFromEmpty   = errors.New("playlist is empty")
	ErrRemovePlaying     = errors.New("this song is playing")
	ErrRemoveNotIn       = errors.New("this song is not in playlist")
	ErrEditCurrent       = errors.New("this is current song")
	ErrLargerTime        = errors.New("time is larger than current song duration")
)

type Status struct {
	Id          uint
	Name        string
	Processing  bool
	Playing     bool
	Time        uint
	CurrentId   uint
	CurrentName string
	Duration    uint
}

type Song struct {
	Id       uint
	Name     string
	Duration uint
	prev     *Song
	next     *Song
}

type Playlist struct {
	Id   uint
	Name string
	sync.RWMutex
	processing bool
	playing    bool
	time       uint
	head       *Song
	tail       *Song
	curr       *Song
	chanPlay   chan struct{}
	chanPaus   chan struct{}
	chanNext   chan struct{}
	chanPrev   chan struct{}
	chanStop   chan struct{}
}

func New(id uint, name string) *Playlist {
	log.Printf("playlist | id %d | created", id)

	return &Playlist{
		Id:         id,
		Name:       name,
		processing: false,
		playing:    false,
		time:       0,
		chanPlay:   make(chan struct{}),
		chanPaus:   make(chan struct{}),
		chanNext:   make(chan struct{}),
		chanPrev:   make(chan struct{}),
		chanStop:   make(chan struct{}),
	}
}

func (pl *Playlist) IsProcessing() bool {
	return pl.processing
}

func (pl *Playlist) IsCurrent(id uint) bool {
	if pl.curr == nil {
		return false
	}

	return pl.curr.Id == id
}

func (pl *Playlist) Process(ctx context.Context) {
	pl.processing = true

	log.Printf("playlist | id %d | active", pl.Id)

	if pl.curr == nil {
		pl.curr = pl.head
	}

	for {
		if err := ctx.Err(); err != nil {
			log.Printf("playlist | id %d | %v", pl.Id, err)

			break
		}

		if !pl.processing || pl.curr == nil {
			log.Printf("playlist | id %d | stopped", pl.Id)

			break
		}

		if !pl.playing {
			pl.processPause()

			continue
		}

		for pl.time <= pl.curr.Duration {
			if ctx.Err() != nil {
				break
			}

			switchRequested := pl.processPlay()

			if switchRequested || !pl.playing || !pl.processing || pl.curr == nil {
				break
			}

			if pl.time == pl.curr.Duration {
				pl.switchNext()

				break
			}

			log.Printf("playlist | id %d | current %d | time %d", pl.Id, pl.curr.Id, pl.time)

			pl.time++
		}
	}

	pl.playing = false
	pl.processing = false

	log.Printf("playlist | id %d | inactive", pl.Id)
}

func (pl *Playlist) processPlay() bool {
	select {
	case <-pl.chanPlay:
		break
	case <-pl.chanPaus:
		pl.playing = false
		break
	case <-pl.chanNext:
		return true
	case <-pl.chanPrev:
		return true
	case <-pl.chanStop:
		break
	default:
		time.Sleep(time.Second)
	}

	return false
}

func (pl *Playlist) processPause() {
	select {
	case <-pl.chanPlay:
		pl.playing = true
		break
	case <-pl.chanPaus:
		break
	case <-pl.chanNext:
		break
	case <-pl.chanPrev:
		break
	case <-pl.chanStop:
		break
	default:
	}
}

func (pl *Playlist) switchNext() {
	pl.curr = pl.curr.next
	pl.time = 0

	if pl.curr != nil {
		log.Printf("playlist | id %d | next | id %d | duration %d", pl.Id, pl.curr.Id, pl.curr.Duration)
	}
}

func (pl *Playlist) Play() error {
	pl.RLock()
	defer pl.RUnlock()

	if !pl.processing {
		return ErrNotProcessed
	}

	if pl.playing {
		return ErrAlreadyPlaying
	}

	pl.chanPlay <- struct{}{}

	pl.playing = true

	log.Printf("playlist | id %d | play | id %d | time %d", pl.Id, pl.curr.Id, pl.time)

	return nil
}

func (pl *Playlist) Pause() error {
	pl.RLock()
	defer pl.RUnlock()

	if !pl.processing {
		return ErrNotProcessed
	}

	if !pl.playing {
		return ErrAlreadyPaused
	}

	pl.chanPaus <- struct{}{}

	pl.playing = false

	log.Printf("playlist | id %d | pause | id %d | time %d", pl.Id, pl.curr.Id, pl.time)

	return nil
}

func (pl *Playlist) Next() error {
	pl.RLock()
	defer pl.RUnlock()

	if !pl.processing {
		return ErrNotProcessed
	}

	if pl.curr.next == nil {
		return ErrSwitchLast
	}

	pl.switchNext()

	pl.chanNext <- struct{}{}

	return nil
}

func (pl *Playlist) Prev() error {
	pl.RLock()
	defer pl.RUnlock()

	if !pl.processing {
		return ErrNotProcessed
	}

	if pl.curr.prev == nil {
		return ErrSwitchFirst
	}

	pl.curr = pl.curr.prev
	pl.time = 0

	log.Printf("playlist | id %d | prev | id %d | duration %d", pl.Id, pl.curr.Id, pl.curr.Duration)

	pl.chanPrev <- struct{}{}

	return nil
}

func (pl *Playlist) Stop() error {
	pl.RLock()
	defer pl.RUnlock()

	if !pl.processing {
		return ErrAlreadyStopped
	}

	pl.chanStop <- struct{}{}

	pl.processing = false

	log.Printf("playlist | id %d | stop", pl.Id)

	return nil
}

func (pl *Playlist) AddSong(id uint, name string, duration uint) error {
	pl.Lock()
	defer pl.Unlock()

	if pl.findSong(id) != nil {
		return ErrSongIdTaken
	}

	song := &Song{
		Id:       id,
		Name:     name,
		Duration: duration,
	}

	if pl.head == nil {
		pl.head = song
		pl.curr = song
	} else {
		song.prev = pl.tail
		pl.tail.next = song
	}

	pl.tail = song

	log.Printf("playlist | id %d | add song | id %d | duration %d", pl.Id, song.Id, song.Duration)

	return nil
}

func (pl *Playlist) Remove(id uint) error {
	pl.Lock()
	defer pl.Unlock()

	if pl.head == nil {
		return ErrRemoveFromEmpty
	}

	if pl.playing && pl.curr.Id == id {
		return ErrRemovePlaying
	}

	var song *Song

	if pl.curr.Id == id {
		song = pl.curr
		pl.time = 0
	} else {
		song = pl.findSong(id)

		if song == nil {
			return ErrRemoveNotIn
		}
	}

	if song == pl.head {
		if song.next == nil {
			pl.head = nil
			pl.tail = nil
			pl.curr = nil
		} else {
			pl.head = song.next
			pl.curr = pl.head
			song.next.prev = nil
		}
	} else if song == pl.tail {
		pl.tail = song.prev
		pl.curr = pl.tail
		song.prev.next = nil
	} else {
		song.prev.next = song.next
		song.next.prev = song.prev
		pl.curr = song.next
	}

	song.next = nil
	song.prev = nil

	log.Printf("playlist | id %d | remove | id %d", pl.Id, song.Id)

	return nil
}

func (pl *Playlist) SetTime(time uint) error {
	pl.RLock()
	defer pl.RUnlock()

	if time > pl.curr.Duration {
		return ErrLargerTime
	}

	pl.time = time

	log.Printf("playlist | id %d | set time | id %d | time %d", pl.Id, pl.curr.Id, pl.time)

	return nil
}

func (pl *Playlist) Status() Status {
	pl.Lock()
	defer pl.Unlock()

	var id uint
	var name string
	var duration uint

	if pl.curr != nil {
		id = pl.curr.Id
		name = pl.curr.Name
		duration = pl.curr.Duration
	}

	log.Printf("playlist | id %d | status | processing %t | playing %t | time %d | id %d | duration %d", pl.Id, pl.processing, pl.playing, pl.time, id, duration)

	return Status{
		Id:          pl.Id,
		Name:        pl.Name,
		Time:        pl.time,
		Playing:     pl.playing,
		Processing:  pl.processing,
		CurrentId:   id,
		CurrentName: name,
		Duration:    duration,
	}
}

func (pl *Playlist) GetSong(id uint) (*Song, error) {
	pl.Lock()
	defer pl.Unlock()

	song := pl.findSong(id)
	if song == nil {
		return nil, ErrSongNotIn
	}

	return song, nil
}

func (pl *Playlist) GetSongsList() []Song {
	pl.Lock()
	defer pl.Unlock()

	var songs []Song

	for s := pl.head; s != nil; s = s.next {
		songs = append(songs, *s)
	}

	return songs
}

func (pl *Playlist) findSong(id uint) *Song {
	var song *Song

	for s := pl.head; s != nil; s = s.next {
		if s.Id == id {
			song = s

			break
		}
	}

	return song
}
