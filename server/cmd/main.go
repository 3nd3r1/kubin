package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/3nd3r1/kubin/server/internal/api"
	"github.com/3nd3r1/kubin/server/internal/blob"
	"github.com/3nd3r1/kubin/server/internal/config"
	"github.com/3nd3r1/kubin/server/internal/server"
	"github.com/3nd3r1/kubin/server/internal/snapshot"
	"github.com/3nd3r1/kubin/server/internal/store"
	"github.com/3nd3r1/kubin/server/internal/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "kubin-server:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := newLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tmpDir := filepath.Join(cfg.DataDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0o750); err != nil {
		return err
	}
	db, err := store.OpenSQLite(ctx, filepath.Join(cfg.DataDir, "kubin.db"))
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	blobs, err := blob.NewFS(filepath.Join(cfg.DataDir, "blobs"))
	if err != nil {
		return err
	}

	srv := &http.Server{
		Handler: api.NewRouter(api.Options{
			Log:            log,
			SecureHeaders:  cfg.Production(),
			Snapshots:      &snapshot.Service{Store: db, Blobs: blobs, TempDir: tmpDir},
			Store:          db,
			Ready:          db.Ping,
			PublicURL:      cfg.PublicURL,
			MaxUploadBytes: cfg.MaxUploadBytes,
			UI:             web.Handler(web.Dist()),
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return err
	}
	log.Info("listening", "addr", ln.Addr().String(), "env", cfg.Env, "data_dir", cfg.DataDir)

	if err := server.Serve(ctx, srv, ln); err != nil {
		return err
	}
	log.Info("stopped")
	return nil
}

func newLogger(level string) *slog.Logger {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: l}))
}
