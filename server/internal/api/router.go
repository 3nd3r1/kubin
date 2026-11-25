// Package api holds the HTTP routes and handlers.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Options struct {
	Log *slog.Logger
	// SecureHeaders adds browser hardening headers to every response.
	SecureHeaders bool

	Snapshots Uploader
	Store     Reader
	Logs      LogBlobs
	// Ready reports whether the database can be used. Nil means always ready.
	Ready func(context.Context) error
	// PublicURL is the base for snapshot links. Empty means taken from the
	// request.
	PublicURL      string
	MaxUploadBytes int64
	// UploadTimeout extends read and write deadlines on the upload handler.
	// Zero leaves the server's ReadTimeout and WriteTimeout in place.
	UploadTimeout time.Duration
	// UploadRate is uploads per client IP per hour. Zero disables rate limiting.
	UploadRate     int
	UploadBurst    int
	TrustedProxies []netip.Prefix
	// Now overrides time.Now for the upload rate limiter. Nil means time.Now.
	Now func() time.Time
	// UI serves every path that isn't an API or health route. Nil means
	// 404 for those paths.
	UI http.Handler
}

type handlers struct {
	opts    Options
	limiter *uploadLimiter
}

func NewRouter(opts Options) http.Handler {
	h := &handlers{
		opts:    opts,
		limiter: newUploadLimiter(opts.UploadRate, opts.UploadBurst, opts.TrustedProxies, opts.Now),
	}
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(requestLog(opts.Log))
	if opts.SecureHeaders {
		r.Use(secureHeaders)
	}

	r.Get("/healthz", ok)
	r.Get("/readyz", h.ready)

	r.Route("/api/v1", func(r chi.Router) {
		r.NotFound(apiNotFound)
		r.MethodNotAllowed(apiMethodNotAllowed)
		r.Post("/snapshots", h.createSnapshot)
		r.Get("/snapshots/{id}", h.getSnapshot)
		r.Get("/snapshots/{id}/resources", h.listResources)
		r.Get("/snapshots/{id}/resources/{kind}/{ns}/{name}", h.getResource)
		r.Get("/snapshots/{id}/logs/{ns}/{pod}/{container}", h.getLog)
	})
	// Unknown API paths are JSON 404s, never the UI.
	r.Handle("/api", http.HandlerFunc(apiNotFound))
	r.Handle("/api/*", http.HandlerFunc(apiNotFound))

	if opts.UI != nil {
		r.Handle("/*", opts.UI)
	}

	return r
}

func apiNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "no such API endpoint")
}

func apiMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *handlers) ready(w http.ResponseWriter, r *http.Request) {
	if h.opts.Ready != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.opts.Ready(ctx); err != nil {
			h.opts.Log.Warn("not ready", "err", err)
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
	}
	ok(w, r)
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func ok(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

func requestLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Debug("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start),
			)
		})
	}
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
