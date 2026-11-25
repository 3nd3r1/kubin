package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/3nd3r1/kubin/server/internal/snapshot"
	"github.com/3nd3r1/kubin/server/internal/store"
)

// Uploader stores and deletes snapshots.
type Uploader interface {
	Upload(ctx context.Context, r io.Reader, ttl time.Duration) (snapshot.UploadResult, error)
	Delete(ctx context.Context, id, token string) error
}

type uploadResponse struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	ExpiresAt   time.Time `json:"expires_at"`
	DeleteToken string    `json:"delete_token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *handlers) createSnapshot(w http.ResponseWriter, r *http.Request) {
	if h.limiter != nil {
		if ok, retryAfter := h.limiter.allow(r); !ok {
			secs, human := formatRetryAfter(retryAfter)
			w.Header().Set("Retry-After", strconv.Itoa(secs))
			writeError(w, http.StatusTooManyRequests, "too many uploads from your address, try again in "+human)
			return
		}
	}

	ttl, err := h.parseExpiresIn(r.URL.Query().Get("expires_in"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.opts.UploadTimeout > 0 {
		deadline := time.Now().Add(h.opts.UploadTimeout)
		rc := http.NewResponseController(w)
		_ = rc.SetReadDeadline(deadline)
		_ = rc.SetWriteDeadline(deadline)
	}

	body := http.MaxBytesReader(w, r.Body, h.opts.MaxUploadBytes)
	res, err := h.opts.Snapshots.Upload(r.Context(), body, ttl)

	var tooBig *http.MaxBytesError
	var unpackedTooBig *snapshot.TooLargeError
	var inv *snapshot.InvalidError
	switch {
	case errors.As(err, &tooBig):
		writeError(w, http.StatusRequestEntityTooLarge, "upload is larger than "+itoa(tooBig.Limit)+" bytes")
		return
	case errors.As(err, &unpackedTooBig):
		writeError(w, http.StatusRequestEntityTooLarge, unpackedTooBig.Error())
		return
	case errors.As(err, &inv):
		writeError(w, http.StatusBadRequest, "invalid snapshot: "+inv.Error())
		return
	case err != nil:
		h.serverError(w, "upload", err)
		return
	}

	url := h.baseURL(r) + "/s/" + res.ID
	h.opts.Log.Info("snapshot uploaded", "id", res.ID, "expires_at", res.ExpiresAt.UTC())
	w.Header().Set("Location", url)
	writeJSON(w, http.StatusCreated, uploadResponse{
		ID:          res.ID,
		URL:         url,
		ExpiresAt:   res.ExpiresAt.UTC(),
		DeleteToken: res.DeleteToken,
	})
}

func (h *handlers) parseExpiresIn(raw string) (time.Duration, error) {
	defTTL := h.opts.SnapshotTTL
	if defTTL <= 0 {
		defTTL = snapshot.DefaultTTL
	}
	maxTTL := h.opts.SnapshotMaxTTL
	if maxTTL <= 0 {
		maxTTL = snapshot.DefaultMaxTTL
	}
	if raw == "" {
		return defTTL, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid expires_in %q: %v", raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("expires_in must be positive (maximum %s)", maxTTL)
	}
	if d > maxTTL {
		return 0, fmt.Errorf("expires_in %s exceeds maximum %s", d, maxTTL)
	}
	return d, nil
}

func (h *handlers) deleteSnapshot(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r.Header)
	if !ok {
		writeError(w, http.StatusForbidden, "missing or invalid delete token")
		return
	}
	id := chi.URLParam(r, "id")
	err := h.opts.Snapshots.Delete(r.Context(), id, token)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "snapshot not found")
		return
	case errors.Is(err, snapshot.ErrInvalidDeleteToken):
		writeError(w, http.StatusForbidden, "invalid delete token")
		return
	case err != nil:
		h.serverError(w, "delete snapshot", err)
		return
	}
	h.opts.Log.Info("snapshot deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

func bearerToken(hdr http.Header) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(hdr.Get("Authorization")), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	if token == "" || strings.ContainsAny(token, " \t") {
		return "", false
	}
	return token, true
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
