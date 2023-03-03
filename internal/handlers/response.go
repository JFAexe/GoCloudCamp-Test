package handlers

import (
	"net/http"

	"gocloudcamp_test/internal/playlist"

	"github.com/go-chi/render"
)

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	Err            error  `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)

	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "invalid request",
		ErrorText:      err.Error(),
	}
}

func ErrInternalError(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "something went wrong",
		ErrorText:      err.Error(),
	}
}

type MsgResponse struct {
	HTTPStatusCode int    `json:"-"`
	MessageText    string `json:"message,omitempty"`
	PlaylistId     uint   `json:"id,omitempty"`
}

func (e *MsgResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)

	return nil
}

type playlistData struct {
	Status playlist.Status `json:"status,omitempty"`
	Songs  []playlist.Song `json:"songs,omitempty"`
}

type PlaylistResponse struct {
	HTTPStatusCode int          `json:"-"`
	Playlist       playlistData `json:"playlist,omitempty"`
}

func (e *PlaylistResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)

	return nil
}

type AllResponse struct {
	HTTPStatusCode int            `json:"-"`
	Playlists      []playlistData `json:"playlists,omitempty"`
}

func (e *AllResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)

	return nil
}
