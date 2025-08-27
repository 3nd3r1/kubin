package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/3nd3r1/kubin/shared/log"
	"github.com/3nd3r1/kubin/api-gateway/internal/config"
	"github.com/3nd3r1/kubin/api-gateway/internal/router"
)

type Server struct {
	httpServer *http.Server
	router     *router.Router
}

func New(router *router.Router) *Server {
	return &Server{
		router: router,
	}
}

func (s *Server) Start(ctx context.Context) error {
	cfg := config.Get()

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           s.withMiddleware(s.router),
		ReadHeaderTimeout: 5 * time.Second, // Defend against slowloris
		ReadTimeout:       time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Trigger graceful shutdown when context is canceled
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.shutdown(shutdownCtx); err != nil {
			log.WithError(err).Error("HTTP server shutdown error")
		}
	}()

	log.With("addr", s.httpServer.Addr).Info("HTTP server starting")

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	log.Info("HTTP server stopped")
	return nil
}

func (s *Server) shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) withMiddleware(h http.Handler) http.Handler {
	cfg := config.Get()

	// Apply common middleware stack
	handler := h

	// Recovery middleware first
	handler = recoveryMiddleware(handler)

	// Then logging
	handler = loggingMiddleware(handler)

	// Then any other middleware
	if cfg.Server.Env == "production" {
		handler = secureHeadersMiddleware(handler)
	}

	return handler
}
