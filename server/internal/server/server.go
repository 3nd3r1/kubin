package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/3nd3r1/kubin/server/internal/config"
	"github.com/3nd3r1/kubin/server/internal/router"
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
    defer s.Shutdown(ctx)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      s.withMiddleware(s.router),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("HTTP server error: %v", err))
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
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
