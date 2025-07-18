package main

import (
	"context"

	"github.com/3nd3r1/kubin/server/internal/api"
	"github.com/3nd3r1/kubin/server/internal/log"
	"github.com/3nd3r1/kubin/server/internal/router"
	"github.com/3nd3r1/kubin/server/internal/server"
)

func main() {
	ctx := context.Background()
	handler := api.NewSnapshotHandler()
	r := router.New()

	router.RegisterRoutes(r, handler)

	srv := server.New(r)
	if err := srv.Start(ctx); err != nil {
		log.WithError(err).Fatal("Failed starting server")
	}
}
