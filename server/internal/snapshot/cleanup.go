package snapshot

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// DeleteExpired removes expired snapshots in batches of CleanupBatch until none
// remain. Stopping the current pass on error keeps a failing blob deletion from
// looping endlessly; the next cleanup tick retries it.
func (s *Service) DeleteExpired(ctx context.Context) (int, error) {
	batch := s.CleanupBatch
	if batch <= 0 {
		batch = DefaultCleanupBatch
	}
	before := s.now().UTC().Round(0)
	deleted := 0
	for {
		if err := ctx.Err(); err != nil {
			return deleted, err
		}
		expired, err := s.Store.ListExpired(ctx, before, batch)
		if err != nil {
			return deleted, err
		}
		if len(expired) == 0 {
			return deleted, nil
		}
		for _, snap := range expired {
			if err := s.deleteSnapshot(ctx, snap.ID, snap.BlobKey, snap.LogsBlobKey); err != nil {
				return deleted, err
			}
			deleted++
		}
		if len(expired) < batch {
			return deleted, nil
		}
	}
}

// RunCleanup deletes expired snapshots on every interval until ctx is cancelled.
func (s *Service) RunCleanup(ctx context.Context, interval time.Duration, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := s.DeleteExpired(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Error("cleanup expired snapshots", "deleted", n, "err", err)
			} else if n > 0 {
				log.Info("cleaned up expired snapshots", "deleted", n)
			}
		}
	}
}
