# Self-hosting

Kubin is one container. It keeps its SQLite database and the uploaded snapshots in `/data`, so all it needs is a volume there.

## docker run

```bash
docker run -d --name kubin -p 8080:8080 -v kubin-data:/data ghcr.io/3nd3r1/kubin
kubin create --server http://localhost:8080
```

Open the printed link. Snapshots stay in the `kubin-data` volume across restarts and upgrades. To start over, remove the container and run `docker volume rm kubin-data`.

The image is built from `scratch` and runs as uid 1001. A named volume gets the right owner on its own. A bind mount needs the directory to be writable by 1001 first, or the server exits with `mkdir /data/tmp: permission denied`:

```bash
mkdir -p ./kubin-data && sudo chown 1001:1001 ./kubin-data
docker run -d -p 8080:8080 -v "$PWD/kubin-data:/data" ghcr.io/3nd3r1/kubin
```

A docker compose file (TASK-27) and the Helm chart (TASK-7) come later. Both run one replica on SQLite and a volume, like `docker run`. One container handles far more than a team or a public instance needs ([capacity.md](capacity.md)), so Postgres and S3 wait until something there runs out (decision-12).

## Behind a proxy

Links handed to the CLI are built from the request: `X-Forwarded-Proto` (or TLS) for the scheme and the `Host` header for the host. If the proxy doesn't pass those through, set `SERVER_PUBLIC_URL` to the address people open, like `https://kubin.example.com`.

Set `SERVER_TRUSTED_PROXIES` to the proxy's CIDR(s) (for example `10.0.0.0/8,172.16.0.0/12`) so upload rate limiting reads the client address from `X-Forwarded-For` instead of counting every request under the proxy's IP. Without it, `X-Forwarded-For` is ignored.

## Configuration

Everything is an env var. Set them with `-e NAME=value`.

| Variable | Default | What it does |
|---|---|---|
| `SERVER_PORT` | `8080` | Port to listen on |
| `SERVER_DATA_DIR` | `/data` in the image, `data` otherwise | Holds `kubin.db`, `blobs/` and `tmp/` for uploads in progress |
| `SERVER_PUBLIC_URL` | empty | Base for the links returned to the CLI. Empty means taken from the request |
| `SERVER_MAX_UPLOAD_BYTES` | `67108864` (64 MiB) | Largest compressed snapshot accepted. If you raise it, raise `kubin create --max-size` too |
| `SERVER_MAX_UNPACKED_BYTES` | `1073741824` (1 GiB) | Largest size a snapshot may unpack to |
| `SERVER_UPLOAD_RATE` | `30` | Uploads allowed per client IP per hour (per `/64` for IPv6). `0` turns the limit off |
| `SERVER_UPLOAD_BURST` | `10` | Burst of uploads allowed before the hourly rate applies |
| `SERVER_TRUSTED_PROXIES` | empty | Comma-separated CIDRs of proxies whose `X-Forwarded-For` header is trusted for rate limiting |
| `SERVER_UPLOAD_TIMEOUT` | `5m` | Read and write deadline for uploads (`POST /api/v1/snapshots`) |
| `SERVER_ENV` | `production` in the image, `development` otherwise | `production` adds browser hardening headers |
| `SERVER_READ_TIMEOUT` | `10s` | Time to read a request (uploads use `SERVER_UPLOAD_TIMEOUT`) |
| `SERVER_WRITE_TIMEOUT` | `10s` | Time to write a response (uploads use `SERVER_UPLOAD_TIMEOUT`) |
| `SERVER_IDLE_TIMEOUT` | `60s` | Keep-alive idle time |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error`. Logs are JSON on stderr |

Durations use Go syntax: `30s`, `2m`. Upload rate limits are kept in memory per replica, so each replica tracks its own counters.

## Health

`GET /healthz` answers `ok` while the process runs. `GET /readyz` also checks the database. The image has no shell or curl, so check them from outside the container.
