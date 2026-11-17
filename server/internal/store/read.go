package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (s *SQL) GetSnapshot(ctx context.Context, id string) (Snapshot, error) {
	var snap Snapshot
	var manifest string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, uploaded_at, captured_at, format_version, manifest, blob_key, size_bytes
		FROM snapshots WHERE id = $1`, id).
		Scan(&snap.ID, &snap.UploadedAt, &snap.CapturedAt, &snap.FormatVersion, &manifest, &snap.BlobKey, &snap.SizeBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return Snapshot{}, fmt.Errorf("snapshot %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return Snapshot{}, err
	}
	snap.Manifest = []byte(manifest)
	return snap, nil
}

func (s *SQL) CountResources(ctx context.Context, id string) ([]KindCount, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT api_group, api_version, kind, COUNT(*)
		FROM resources WHERE snapshot_id = $1
		GROUP BY api_group, api_version, kind
		ORDER BY api_group, api_version, kind`, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []KindCount{}
	for rows.Next() {
		var c KindCount
		if err := rows.Scan(&c.Group, &c.Version, &c.Kind, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

const resourceColumns = `api_group, api_version, kind, namespace, name, object`

func scanResource(sc interface{ Scan(...any) error }) (Resource, error) {
	var r Resource
	var obj string
	if err := sc.Scan(&r.Group, &r.Version, &r.Kind, &r.Namespace, &r.Name, &obj); err != nil {
		return Resource{}, err
	}
	r.Object = []byte(obj)
	return r, nil
}

func (s *SQL) ListResources(ctx context.Context, id string, f Filter) ([]Resource, error) {
	where := []string{"snapshot_id = $1"}
	args := []any{id}
	add := func(col, val string) {
		args = append(args, val)
		where = append(where, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	if f.Group != "" {
		add("api_group", f.Group)
	}
	if f.Kind != "" {
		add("kind", f.Kind)
	}
	switch {
	case f.ClusterScoped:
		add("namespace", "")
	case f.Namespace != "":
		add("namespace", f.Namespace)
	}

	rows, err := s.db.QueryContext(ctx, `SELECT `+resourceColumns+` FROM resources
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY api_group, api_version, kind, namespace, name`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Resource{}
	for rows.Next() {
		r, err := scanResource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQL) GetResource(ctx context.Context, id string, k Key) (Resource, error) {
	r, err := scanResource(s.db.QueryRowContext(ctx, `SELECT `+resourceColumns+` FROM resources
		WHERE snapshot_id = $1 AND kind = $2 AND namespace = $3 AND name = $4
		  AND ($5 = '' OR api_group = $5)
		ORDER BY api_group, api_version
		LIMIT 1`, id, k.Kind, k.Namespace, k.Name, k.Group))
	if errors.Is(err, sql.ErrNoRows) {
		return Resource{}, fmt.Errorf("%s %s/%s: %w", k.Kind, k.Namespace, k.Name, ErrNotFound)
	}
	return r, err
}
