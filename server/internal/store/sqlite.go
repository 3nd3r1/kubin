package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	_ "modernc.org/sqlite" // pure Go, no CGO
)

// OpenSQLite opens or creates the database file at path and migrates it.
func OpenSQLite(ctx context.Context, path string) (*SQL, error) {
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "synchronous(NORMAL)")
	q.Set("_txlock", "immediate")
	dsn := "file:" + path + "?" + q.Encode()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := migrate(ctx, db, "sqlite"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate %s: %w", path, err)
	}
	// SQLite allows one writer. A second pool with one connection queues
	// writers in Go, where they wait on their context, instead of inside
	// SQLite, where they give up with SQLITE_BUSY after busy_timeout.
	wdb, err := sql.Open("sqlite", dsn)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	wdb.SetMaxOpenConns(1)
	return &SQL{db: db, wdb: wdb}, nil
}
