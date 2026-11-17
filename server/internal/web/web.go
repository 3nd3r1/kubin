// Package web serves the UI from ui/, built into dist/ and embedded in the
// binary.
package web

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

//go:embed all:dist
var dist embed.FS

// Dist is the built UI. It holds only a placeholder until `make build` runs
// in ui/.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err) // the embed pattern guarantees the directory
	}
	return sub
}

const notBuilt = "The UI is not built into this binary. Run `make build` in ui/ and rebuild the server.\n"

// Handler serves files from fsys. Paths without a file get index.html so
// the client-side router can handle them, except under /assets/, where a
// missing file is a 404. Hashed assets are cached for good; index.html is
// revalidated on every load so a new release shows up.
func Handler(fsys fs.FS) http.Handler {
	index, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		index = nil
	}
	started := time.Now()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name != "" && name != "index.html" {
			if st, err := fs.Stat(fsys, name); err == nil && !st.IsDir() {
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				http.ServeFileFS(w, r, fsys, name)
				return
			} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if strings.HasPrefix(name, "assets/") {
				http.NotFound(w, r)
				return
			}
		}

		if index == nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(notBuilt))
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", started, bytes.NewReader(index))
	})
}
