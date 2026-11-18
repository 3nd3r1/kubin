# One image: the UI built into the server binary, on scratch.
# docker build -t kubin .

FROM node:22-alpine AS ui

WORKDIR /src/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci

COPY ui/ ./
# Writes ../server/internal/web/dist (vite.config.ts).
RUN npm run build

FROM golang:1.26-alpine AS server

WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
COPY --from=ui /src/server/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/kubin-server ./cmd
# scratch has no mkdir, so the data directory is made here.
RUN mkdir -p /out/data

FROM scratch

COPY --from=server /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=server /out/kubin-server /kubin-server
COPY --from=server --chown=1001:1001 /out/data /data

LABEL org.opencontainers.image.title="kubin" \
      org.opencontainers.image.description="Pastebin for Kubernetes clusters" \
      org.opencontainers.image.source="https://github.com/3nd3r1/kubin"

ENV SERVER_DATA_DIR=/data \
    SERVER_ENV=production

USER 1001:1001
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/kubin-server"]
