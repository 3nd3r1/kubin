CREATE TABLE snapshots (
    id             TEXT PRIMARY KEY,
    uploaded_at    TIMESTAMP NOT NULL,
    captured_at    TIMESTAMP NOT NULL,
    format_version INTEGER NOT NULL,
    manifest       TEXT NOT NULL,
    blob_key       TEXT NOT NULL,
    size_bytes     BIGINT NOT NULL
);

CREATE TABLE resources (
    snapshot_id TEXT NOT NULL REFERENCES snapshots (id) ON DELETE CASCADE,
    api_group   TEXT NOT NULL,
    api_version TEXT NOT NULL,
    kind        TEXT NOT NULL,
    namespace   TEXT NOT NULL,
    name        TEXT NOT NULL,
    object      TEXT NOT NULL,
    PRIMARY KEY (snapshot_id, api_group, api_version, kind, namespace, name)
);

CREATE INDEX resources_by_kind ON resources (snapshot_id, kind, namespace);
