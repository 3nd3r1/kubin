# Hosted instance (`kubin.ender.fi`)

The public Kubin instance runs as a single container on Coolify with SQLite and a named volume at `/data` (decision-12).

## What runs where

- **Host:** Coolify server behind its built-in Traefik proxy with Let's Encrypt TLS.
- **Domain:** `https://kubin.ender.fi`
- **Image:** `ghcr.io/3nd3r1/kubin:sha-<commit>` (pinned commit tags built on push to `main`, never `latest`).
- **Storage:** One Docker named volume mounted at `/data` (holds `kubin.db`, `blobs/`, and `tmp/`). A named volume inherits uid `1001:1001` from the image; if you ever switch to a bind mount, run `chown 1001:1001` on the host directory first.

## Coolify resource settings

Create a **Docker Image** application in Coolify:

- **Image:** `ghcr.io/3nd3r1/kubin`
- **Tag:** current release tag (e.g. `v0.1.0`)
- **Port:** `8080`
- **Domain:** `https://kubin.ender.fi`
- **Persistent Storage (Volume):**
  - Name: `kubin-data`
  - Destination path: `/data`
- **Environment variables:**

| Variable | Value | Why |
|---|---|---|
| `SERVER_PUBLIC_URL` | `https://kubin.ender.fi` | Canonical base URL for links returned to the CLI |
| `SERVER_TRUSTED_PROXIES` | `<coolify-proxy-cidr>` (typically `10.0.0.0/8,172.16.0.0/12`) | Lets upload rate limiting read client IPs from Traefik's `X-Forwarded-For` |
| `SERVER_MAX_UPLOAD_BYTES` | `67108864` | 64 MiB compressed snapshot cap |
| `SERVER_MAX_UNPACKED_BYTES` | `1073741824` | 1 GiB unpacked snapshot cap |
| `SERVER_UPLOAD_RATE` | `30` | Uploads per client IP per hour |
| `SERVER_UPLOAD_BURST` | `10` | Burst allowance |
| `SERVER_UPLOAD_TIMEOUT` | `5m` | Read/write deadline for uploads |
| `SERVER_SNAPSHOT_TTL` | `168h` | Default 7-day expiry |
| `SERVER_SNAPSHOT_MAX_TTL` | `720h` | Maximum 30-day expiry via `--expire` |
| `SERVER_CLEANUP_INTERVAL` | `10m` | Expired snapshot cleanup cadence |
| `LOG_LEVEL` | `info` | JSON logs on stderr |

### Traefik timeouts

Coolify's proxy is Traefik. Verify that the `https` entrypoint's `respondingTimeouts.readTimeout` (and any middleware body-size limit) allows a 64 MiB upload over a slow connection to finish within `SERVER_UPLOAD_TIMEOUT` (`5m`). Traefik v3 defaults `readTimeout` to `60s`, which can cut off slow uploads unless raised to at least `300s` (or `0s`).

## Disk and backups

- **Sizing:** A typical snapshot is 100 KB–2 MB compressed; worst-case 64 MiB. With a 7-day default TTL and 30 uploads/IP/hour, a 20–50 GB volume gives wide headroom.
- **Monitoring:** Watch host disk usage in Coolify's server metrics or with a host alert (e.g. alert at 80% full on the Docker volume partition). If disk ever fills, lower `SERVER_SNAPSHOT_TTL` and `SERVER_SNAPSHOT_MAX_TTL` or prune old snapshots.
- **Backups:** Optional for a pastebin whose links expire after 7 days (decision-12). If you want point-in-time backups of `kubin.db` later, Litestream to S3-compatible storage can run alongside the volume without changing the server.

## Secrets

No secrets are stored in git. All deploy credentials live in **GitHub repository Settings → Secrets and variables → Actions**:

| Secret | Where it comes from | Used by |
|---|---|---|
| `COOLIFY_URL` | Base URL of your Coolify instance (e.g. `https://coolify.example.com`) | `.github/workflows/deploy.yml` |
| `COOLIFY_APP_UUID` | Application UUID from Coolify's URL / resource settings | `.github/workflows/deploy.yml` |
| `COOLIFY_API_TOKEN` | Coolify → Keys & Tokens → API Tokens (deploy + update permissions) | `.github/workflows/deploy.yml` |

## Deploying and rolling back

### Automatic (CD on `main`)

Pushing to `main` with changes under `server/`, `ui/`, or `Dockerfile` triggers `.github/workflows/deploy.yml`. It builds and pushes `ghcr.io/3nd3r1/kubin:sha-<short-sha>`, updates the Coolify application's image tag, triggers a deploy, and polls `https://kubin.ender.fi/readyz` until it returns `ok`.

Cutting a `v*` tag runs `.github/workflows/release.yml` to publish CLI binaries and the multi-arch self-hosted image (`ghcr.io/3nd3r1/kubin:<tag>` and `latest`).

### Manual / Rollback

- **Via GitHub Actions:** Run the **Deploy** workflow manually (`workflow_dispatch`). Leave `tag` empty to build and deploy current `main`, or pass an existing tag (e.g. `sha-cddaaad` or `v0.1.0`) to roll back without rebuilding.
- **Via Coolify UI:** Open the Kubin application, change the Docker image tag to a previous `sha-*` or `v*` tag, and click **Redeploy**. The `/data` named volume is preserved across redeploys.

## One-time setup checklist (human steps)

1. **DNS:** Add an `A` (and `AAAA` if applicable) record for `kubin.ender.fi` pointing to the Coolify server.
2. **Coolify resource:** Create the Docker Image application (`ghcr.io/3nd3r1/kubin`, port `8080`, domain `https://kubin.ender.fi`, named volume `kubin-data` at `/data`, env vars from the table above).
3. **Traefik timeout:** Confirm Traefik's `https` entrypoint `readTimeout` is at least `300s` (or `0s`) for large uploads.
4. **Coolify API token:** Create an API token in Coolify and note the application UUID.
5. **GitHub secrets:** Add `COOLIFY_URL`, `COOLIFY_APP_UUID`, and `COOLIFY_API_TOKEN` in GitHub Actions secrets.
6. **Verify:**
   - `curl -fsS https://kubin.ender.fi/readyz` prints `ok`
   - `kubin create --server https://kubin.ender.fi` against a kind cluster prints a link that opens
   - Redeploy in Coolify and verify the snapshot link still loads
