# Self-hosting

Kubin is one container. It keeps its SQLite database and the uploaded snapshots in `/data`, so all it needs is a volume there.

## docker run

```bash
docker build -t kubin .                                  # from the repo root, until images are released (TASK-19)
docker run -d --name kubin -p 8080:8080 -v kubin-data:/data kubin
kubin create --server http://localhost:8080
```

Open the printed link. Snapshots stay in the `kubin-data` volume across restarts and upgrades. To start over, remove the container and run `docker volume rm kubin-data`.

The image is built from `scratch` and runs as uid 1001. A named volume gets the right owner on its own. A bind mount needs the directory to be writable by 1001 first, or the server exits with `mkdir /data/tmp: permission denied`:

```bash
mkdir -p ./kubin-data && sudo chown 1001:1001 ./kubin-data
docker run -d -p 8080:8080 -v "$PWD/kubin-data:/data" kubin
```

Docker compose with Postgres and MinIO (TASK-27) and the Helm chart (TASK-7) come later. With SQLite and a local directory, run one replica.

## Behind a proxy

Links handed to the CLI are built from the request: `X-Forwarded-Proto` (or TLS) for the scheme and the `Host` header for the host. If the proxy doesn't pass those through, set `SERVER_PUBLIC_URL` to the address people open, like `https://kubin.example.com`.

## Configuration

Everything is an env var. Set them with `-e NAME=value`.

| Variable | Default | What it does |
|---|---|---|
| `SERVER_PORT` | `8080` | Port to listen on |
| `SERVER_DATA_DIR` | `/data` in the image, `data` otherwise | Holds `kubin.db`, `blobs/` and `tmp/` for uploads in progress |
| `SERVER_PUBLIC_URL` | empty | Base for the links returned to the CLI. Empty means taken from the request |
| `SERVER_MAX_UPLOAD_BYTES` | `67108864` (64 MiB) | Largest snapshot accepted. If you raise it, raise `kubin create --max-size` too |
| `SERVER_ENV` | `production` in the image, `development` otherwise | `production` adds browser hardening headers |
| `SERVER_READ_TIMEOUT` | `10s` | Time to read a whole request, upload included |
| `SERVER_WRITE_TIMEOUT` | `10s` | Time to write a response |
| `SERVER_IDLE_TIMEOUT` | `60s` | Keep-alive idle time |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error`. Logs are JSON on stderr |

Durations use Go syntax: `30s`, `2m`.

## Health

`GET /healthz` answers `ok` while the process runs. `GET /readyz` also checks the database. The image has no shell or curl, so check them from outside the container.
