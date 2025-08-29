FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY ../go.mod ../go.sum ./

RUN go mod download

# Copy source code
COPY ../ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o api-gateway \
    ./cmd/

FROM scratch

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Create non-root user directories
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
COPY --from=builder /app/api-gateway /api-gateway

# Metadata labels following OCI standards
LABEL org.opencontainers.image.title="Kubin API Gateway"
LABEL org.opencontainers.image.description="API Gateway service for Kubin - Kubernetes cluster pastebin"
LABEL org.opencontainers.image.vendor="Kubin"
LABEL org.opencontainers.image.source="https://github.com/3nd3r1/kubin"
LABEL org.opencontainers.image.documentation="https://github.com/3nd3r1/kubin/tree/main/services/api-gateway"

# Security: Run as non-root user
# Create a user in the builder stage and copy to scratch
USER 1001:1001

# Expose the default port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD ["/api-gateway", "healthcheck"]

# Set the entrypoint
ENTRYPOINT ["/api-gateway"]
