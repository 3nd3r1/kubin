package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/3nd3r1/kubin/server/internal/snapshot"
)

// Uploader stores an uploaded snapshot and returns its ID.
type Uploader interface {
	Upload(ctx context.Context, r io.Reader) (string, error)
}

type uploadResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *handlers) createSnapshot(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, h.opts.MaxUploadBytes)
	id, err := h.opts.Snapshots.Upload(r.Context(), body)

	var tooBig *http.MaxBytesError
	var inv *snapshot.InvalidError
	switch {
	case errors.As(err, &tooBig):
		writeError(w, http.StatusRequestEntityTooLarge, "upload is larger than "+itoa(tooBig.Limit)+" bytes")
		return
	case errors.As(err, &inv):
		writeError(w, http.StatusBadRequest, "invalid snapshot: "+inv.Error())
		return
	case err != nil:
		h.serverError(w, "upload", err)
		return
	}

	url := h.baseURL(r) + "/s/" + id
	h.opts.Log.Info("snapshot uploaded", "id", id)
	w.Header().Set("Location", url)
	writeJSON(w, http.StatusCreated, uploadResponse{ID: id, URL: url})
}

// baseURL is the configured public URL, or else the scheme and host the
// request came in on.
func (h *handlers) baseURL(r *http.Request) string {
	if h.opts.PublicURL != "" {
		return strings.TrimRight(h.opts.PublicURL, "/")
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p == "http" || p == "https" {
		scheme = p
	}
	return scheme + "://" + r.Host
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
