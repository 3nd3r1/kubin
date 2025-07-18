package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/3nd3r1/kubin/server/internal/log"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap the response writer to capture status code
        wrapped := wrapResponseWriter(w)
        
        defer func() {
            log.With("method", r.Method).
                With("path", r.URL.Path).
                With("status", wrapped.status).
                With("remote", r.RemoteAddr).
                With("duration", time.Since(start)).
                Debug("request received")
        }()
        
        next.ServeHTTP(wrapped, r)
    })
}

func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Error(fmt.Sprintf("panic: %v", err))
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}

func secureHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        next.ServeHTTP(w, r)
    })
}

// responseWriter wrapper to capture status code
type responseWriter struct {
    http.ResponseWriter
    status int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
    return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}
