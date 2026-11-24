package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations
var migrations embed.FS

// SQL is a Store on database/sql. The dialect only picks the migrations.
type SQL struct {
	db  *sql.DB // reads
	wdb *sql.DB // writes; may be the same pool as db
}

// migrate applies every migration in migrations/<dialect> that isn't recorded
// in schema_migrations yet, each in its own transaction.
func migrate(ctx context.Context, db *sql.DB, dialect string) error {
	if _, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[int]bool{}
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			_ = rows.Close()
			return err
		}
		applied[v] = true
	}
	if err := rows.Close(); err != nil {
		return err
	}

	dir := path.Join("migrations", dialect)
	entries, err := fs.ReadDir(migrations, dir)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		version, err := strconv.Atoi(strings.SplitN(e.Name(), "_", 2)[0])
		if err != nil {
			return fmt.Errorf("migration %s: name must start with a number", e.Name())
		}
		if applied[version] {
			continue
		}
		body, err := fs.ReadFile(migrations, path.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		if err := applyMigration(ctx, db, version, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", e.Name(), err)
		}
	}
	return nil
}

func applyMigration(ctx context.Context, db *sql.DB, version int, body string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, body); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQL) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *SQL) Close() error {
	err := s.wdb.Close()
	if s.db != s.wdb {
		if err2 := s.db.Close(); err == nil {
			err = err2
		}
	}
	return err
}

func (s *SQL) BeginSnapshot(ctx context.Context, snap Snapshot) (SnapshotTx, error) {
	tx, err := s.wdb.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO snapshots (id, uploaded_at, captured_at, format_version, manifest, blob_key, size_bytes, logs_blob_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		snap.ID, snap.UploadedAt.UTC(), snap.CapturedAt.UTC(), snap.FormatVersion,
		string(snap.Manifest), snap.BlobKey, snap.SizeBytes, snap.LogsBlobKey)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("insert snapshot: %w", err)
	}
	insert, err := tx.PrepareContext(ctx, `
		INSERT INTO resources (snapshot_id, api_group, api_version, kind, namespace, name, object)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	insertLog, err := tx.PrepareContext(ctx, `
		INSERT INTO logs (snapshot_id, namespace, pod, container, previous, byte_offset, byte_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		_ = insert.Close()
		_ = tx.Rollback()
		return nil, err
	}
	return &sqlSnapshotTx{tx: tx, insert: insert, insertLog: insertLog, id: snap.ID}, nil
}

type sqlSnapshotTx struct {
	tx        *sql.Tx
	insert    *sql.Stmt
	insertLog *sql.Stmt
	id        string
}

func (t *sqlSnapshotTx) AddLog(ctx context.Context, l Log) error {
	_, err := t.insertLog.ExecContext(ctx, t.id, l.Namespace, l.Pod, l.Container, boolInt(l.Previous), l.Offset, l.Size)
	if err != nil {
		return fmt.Errorf("insert log %s/%s %s: %w", l.Namespace, l.Pod, l.Container, err)
	}
	return nil
}

// boolInt stores booleans as 0 and 1, which every dialect reads the same.
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (t *sqlSnapshotTx) AddResource(ctx context.Context, r Resource) error {
	_, err := t.insert.ExecContext(ctx, t.id, r.Group, r.Version, r.Kind, r.Namespace, r.Name, string(r.Object))
	if err != nil {
		return fmt.Errorf("insert %s %s/%s: %w", r.Kind, r.Namespace, r.Name, err)
	}
	return nil
}

func (t *sqlSnapshotTx) Commit() error {
	_ = t.insert.Close()
	_ = t.insertLog.Close()
	return t.tx.Commit()
}

func (t *sqlSnapshotTx) Rollback() error {
	_ = t.insert.Close()
	_ = t.insertLog.Close()
	if err := t.tx.Rollback(); err != nil && err != sql.ErrTxDone {
		return err
	}
	return nil
}
