package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/3nd3r1/kubin/api-gateway/internal/api"
	"github.com/3nd3r1/kubin/api-gateway/internal/router"
	"github.com/3nd3r1/kubin/api-gateway/internal/server"
	"github.com/3nd3r1/kubin/shared/log"
)

func main() {
	// Create context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize logger
	log.Info("Starting Kubin server...")

	// Create handlers
	handler := api.NewSnapshotHandler()

	// Create router and register routes
	r := router.New()
	router.RegisterRoutes(r, handler)

	// Create and start server
	srv := server.New(r)

	log.Info("Server initialized, starting...")
	if err := srv.Start(ctx); err != nil {
		log.WithError(err).Fatal("Server failed")
	}
}
