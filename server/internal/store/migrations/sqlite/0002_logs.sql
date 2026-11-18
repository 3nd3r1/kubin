-- Logs are copied out of the tarball at upload into one blob per snapshot.
-- Each row says where one log sits in that blob.
ALTER TABLE snapshots ADD COLUMN logs_blob_key TEXT NOT NULL DEFAULT '';

CREATE TABLE logs (
    snapshot_id TEXT NOT NULL REFERENCES snapshots (id) ON DELETE CASCADE,
    namespace   TEXT NOT NULL,
    pod         TEXT NOT NULL,
    container   TEXT NOT NULL,
    previous    INTEGER NOT NULL,
    byte_offset BIGINT NOT NULL,
    byte_size   BIGINT NOT NULL,
    PRIMARY KEY (snapshot_id, namespace, pod, container, previous)
);
