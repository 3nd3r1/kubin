FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/kubin-server ./cmd

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/kubin-server /kubin-server

LABEL org.opencontainers.image.title="kubin" \
      org.opencontainers.image.description="Pastebin for Kubernetes clusters" \
      org.opencontainers.image.source="https://github.com/3nd3r1/kubin"

USER 1001:1001
EXPOSE 8080
ENTRYPOINT ["/kubin-server"]
