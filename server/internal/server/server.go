// Package server runs the HTTP server until its context ends.
package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// ShutdownTimeout is how long in-flight requests get to finish after the
// context ends.
const ShutdownTimeout = 10 * time.Second

// Serve serves srv on ln until ctx is done, then shuts down gracefully and
// returns once in-flight requests have finished.
func Serve(ctx context.Context, srv *http.Server, ln net.Listener) error {
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
