package api

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/3nd3r1/kubin/server/internal/store"
)

// maxLogLine is the longest line the q filter looks at. Longer lines are
// matched and returned cut to this length. The CLI caps a whole log at 1 MiB,
// so only hand-made snapshots get here.
const maxLogLine = 1 << 20

// LogBlobs reads ranges of a snapshot's logs blob. blob.Store implements it.
type LogBlobs interface {
	OpenRange(ctx context.Context, key string, offset, length int64) (io.ReadCloser, error)
}

// getLog streams one container log as plain text. previous=true picks the
// run before the last restart. q keeps only the lines that contain it,
// ignoring ASCII case. Nothing is held in memory beyond one line.
func (h *handlers) getLog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	previous := false
	if p := q.Get("previous"); p != "" {
		var err error
		if previous, err = strconv.ParseBool(p); err != nil {
			writeError(w, http.StatusBadRequest, "previous must be true or false")
			return
		}
	}
	snap, ok := h.snapshot(w, r)
	if !ok {
		return
	}
	l, err := h.opts.Store.GetLog(r.Context(), snap.ID, store.LogKey{
		Namespace: chi.URLParam(r, "ns"),
		Pod:       chi.URLParam(r, "pod"),
		Container: chi.URLParam(r, "container"),
		Previous:  previous,
	})
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "log not found")
		return
	}
	if err != nil {
		h.serverError(w, "get log", err)
		return
	}
	rc, err := h.opts.Logs.OpenRange(r.Context(), snap.LogsBlobKey, l.Offset, l.Size)
	if err != nil {
		h.serverError(w, "open log", err)
		return
	}
	defer func() { _ = rc.Close() }()

	// Logs are whatever the containers wrote: never let a browser treat
	// them as anything but text.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=3600")

	search := q.Get("q")
	if search == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(l.Size, 10))
		if _, err := io.Copy(w, rc); err != nil {
			h.opts.Log.Warn("stream log", "err", err)
		}
		return
	}
	if err := grepLines(w, rc, []byte(search)); err != nil {
		h.opts.Log.Warn("search log", "err", err)
	}
}

// grepLines writes the lines of r that contain q, ignoring ASCII case. It
// reuses two line buffers, so a long log costs no more than its longest
// line.
func grepLines(w io.Writer, r io.Reader, q []byte) error {
	q = lowerASCII(nil, q)
	br := bufio.NewReaderSize(r, 64*1024)
	var line, lower []byte
	cut := false
	for {
		chunk, err := br.ReadSlice('\n')
		if room := maxLogLine - len(line); len(chunk) > room {
			chunk, cut = chunk[:room], true
		}
		line = append(line, chunk...)
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		lower = lowerASCII(lower[:0], line)
		if len(line) > 0 && bytes.Contains(lower, q) {
			if cut && line[len(line)-1] != '\n' {
				line = append(line, '\n')
			}
			if _, werr := w.Write(line); werr != nil {
				return werr
			}
		}
		line, cut = line[:0], false
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// lowerASCII appends b to dst with A-Z lowered.
func lowerASCII(dst, b []byte) []byte {
	for _, c := range b {
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		dst = append(dst, c)
	}
	return dst
}
