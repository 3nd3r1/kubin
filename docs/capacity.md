# Capacity

How much one kubin container handles, measured. decision-12 rests on these numbers: one node with SQLite and a local directory is enough for m-3 and the public instance, and Postgres and S3 wait until something here runs out.

## Setup

Measured on 2026-11-24 at commit 985bceb.

- Laptop: Intel i5-1135G7 (4 cores, 8 threads), 16 GB RAM, NVMe disk. Client and server on the same machine, so no network in between.
- The server binary built from `server/`, SQLite and a local directory (the `docker run` defaults), `SERVER_READ_TIMEOUT` and `SERVER_WRITE_TIMEOUT` at 10m so the timeouts don't cut uploads.
- Three snapshots:

| Snapshot | Resources | Log files | Unpacked | Gzipped |
|---|---|---|---|---|
| real | 61 | 24 | 0.6 MB | 70 KB |
| medium | 601 | 240 | 14 MB | 3.3 MB |
| big | 2,401 | 960 | 204 MB | 52 MB |

`real` is `kubin create` against the demo cluster (`demo/setup.sh`). `medium` and `big` are that snapshot cloned 10 and 40 times by `tools/bench/gen.py`, with 50 KiB and 200 KiB of generated log lines per file. The generated lines compress worse than real logs, so `big` sits close to the 64 MiB upload limit and is a worst case.

## Results

Uploads, `go run load.go upload`:

| Snapshot | Concurrent | Uploads/s | p50 | p95 | Failed |
|---|---|---|---|---|---|
| real | 1 | 68 | 14 ms | 19 ms | 0 of 50 |
| real | 8 | 59 | 16 ms | 0.7 s | 0 of 200 |
| medium | 1 | 3.5 | 0.28 s | 0.31 s | 0 of 10 |
| medium | 8 | 3.9 | 0.34 s | 5.0 s | 7 of 40 |
| big | 1 | 0.27 | 3.7 s | 3.9 s | 0 of 3 |
| big | 4 | 0.53 | 5.2 s | 7.5 s | 4 of 8 |

Reads, `go run load.go read`: 32 clients for 15 s, rotating over what the UI asks for (the snapshot, the resource list, every 7th resource, a log).

| Snapshot | Requests/s | p50 | p95 | p99 |
|---|---|---|---|---|
| medium | 2,200 | 11 ms | 34 ms | 74 ms |
| big | 1,460 | 19 ms | 44 ms | 63 ms |
| medium, while 2 big uploads run | 2,290 | 11 ms | 34 ms | 74 ms |

Memory peaked at 400 MB during the concurrent big uploads and sat around 75 MB otherwise. After about 300 uploads, mostly big ones, the data directory held 2.6 GB, of which 173 MB was the SQLite file.

## What it means

- **Uploads.** One at a time, an upload takes as long as writing its resource index. That's about 10,000 medium snapshots an hour, or 1,000 of the worst case. The public instance limits each address to 30 an hour (TASK-22).
- **Reads.** Opening a snapshot in the UI takes roughly 10 to 50 requests, so 1,500 requests a second is 30 to 150 people opening links every second. Uploads don't slow reads down, because SQLite in WAL mode lets reads run during a write.
- **Disk is the limit that matters.** Storage is about uploads a day × snapshot size × days kept. 200 uploads a day of 3 MB kept for 7 days is about 4 GB. Expiry (TASK-21) keeps it bounded.
- A small VPS with 2 to 4 vCPUs and slower disks may be 2 to 4 times slower than this laptop. That still leaves two to three orders of magnitude of headroom over the expected load.
- Postgres and S3 would buy availability across replicas, not capacity at this scale. Revisit them when uploads queue up behind each other in normal use, when the disk can't hold a week of snapshots, or when someone needs the instance to survive a node going down.

## Fixed: SQLITE_BUSY on concurrent uploads

The failed uploads in the first table were 500s with `database is locked (5) (SQLITE_BUSY)`. Each upload holds SQLite's write lock for its whole index transaction, and concurrent uploads waited inside SQLite and gave up after `busy_timeout(5000)`. TASK-48 gives writes their own pool of one connection, so uploads queue in Go. Re-run on 2026-11-24: 40 medium uploads at 8 concurrent, 0 failed, no SQLITE_BUSY in the server log (p50 0.41 s, p95 1.4 s). That run used a synthetic 600-resource snapshot (smaller logs, 3.2 MB unpacked) because no demo cluster was at hand, so the timings are not comparable with the table above. The big case was not re-run.

## Known problems

- **The resource list is not paginated.** For `big`, `GET /api/v1/snapshots/{id}/resources` returns 5.4 MB of JSON in one response. Reads are fast anyway, but the UI downloads all of it.

## Running it again

Build the tools and the snapshots:

```sh
cd cli && go build -o /tmp/kubin . && cd ..
cd server && go build -o /tmp/kubin-server ./cmd && cd ..
/tmp/kubin create -o /tmp/real.tar.gz          # against the demo cluster
mkdir -p /tmp/real && tar xzf /tmp/real.tar.gz -C /tmp/real
python3 tools/bench/gen.py /tmp/real /tmp/medium.tar.gz 10 50
python3 tools/bench/gen.py /tmp/real /tmp/big.tar.gz 40 200
```

Start a server on an empty data directory:

```sh
SERVER_PORT=18080 SERVER_DATA_DIR=/tmp/kubin-bench SERVER_READ_TIMEOUT=10m SERVER_WRITE_TIMEOUT=10m \
  LOG_LEVEL=warn /tmp/kubin-server
```

Then, from `tools/bench`:

```sh
go run load.go upload http://localhost:18080 /tmp/medium.tar.gz 8 40
go run load.go read http://localhost:18080 <snapshot id> 32 15 checkout/<pod>/<container>
```

`upload` prints the ID of the last snapshot it stored, which `read` takes. The pod and container must have a log in the snapshot; `manifest.json` lists them. Write new results into the tables above with the date and commit.
