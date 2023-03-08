package handlers

import (
	"net/http"

	"gocloudcamp_test/internal/playlist"

	"github.com/go-chi/render"
)

type errorResponse struct {
	HTTPStatusCode int    `json:"-"`
	MessageText    string `json:"message,omitempty"`
	ErrorText      string `json:"error,omitempty"`
}

func (er *errorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, er.HTTPStatusCode)

	return nil
}

var (
	responseNotFound = &errorResponse{
		HTTPStatusCode: http.StatusNotFound,
		MessageText:    "invalid request",
		ErrorText:      "route does not exist",
	}
	responseNotAllowed = &errorResponse{
		HTTPStatusCode: http.StatusMethodNotAllowed,
		MessageText:    "invalid request",
		ErrorText:      "method is not valid",
	}
)

func responseInvalidRequest(err error) render.Renderer {
	return &errorResponse{
		HTTPStatusCode: http.StatusBadRequest,
		MessageText:    "invalid request",
		ErrorText:      err.Error(),
	}
}

func responseInternalError(err error) render.Renderer {
	return &errorResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		MessageText:    "something went wrong",
		ErrorText:      err.Error(),
	}
}

type messageResponse struct {
	HTTPStatusCode int    `json:"-"`
	MessageText    string `json:"message,omitempty"`
	PlaylistId     uint   `json:"id,omitempty"`
}

func (mr *messageResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, mr.HTTPStatusCode)

	return nil
}

type playlistData struct {
	Status playlist.Status `json:"status,omitempty"`
	Songs  []playlist.Song `json:"songs,omitempty"`
}

type playlistResponse struct {
	HTTPStatusCode int          `json:"-"`
	Playlist       playlistData `json:"playlist,omitempty"`
}

func (pr *playlistResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, pr.HTTPStatusCode)

	return nil
}

type allResponse struct {
	HTTPStatusCode int            `json:"-"`
	Playlists      []playlistData `json:"playlists,omitempty"`
}

func (ar *allResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, ar.HTTPStatusCode)

	return nil
}
