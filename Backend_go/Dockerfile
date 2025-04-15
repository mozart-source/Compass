# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make gcc musl-dev && \
    update-ca-certificates

WORKDIR /build

# Copy go mod files first for better cache utilization
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with security flags and optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=docker-build -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -trimpath \
    -o main ./cmd/api

# Final stage
FROM alpine:3.19

# Add security updates and basic tools
RUN apk --no-cache add ca-certificates tzdata curl && \
    update-ca-certificates && \
    adduser -D -g '' appuser

WORKDIR /app

# Create all required directories and set ownership before copying anything
RUN mkdir -p /app/config /app/certs /app/logs /app/data && \
    chown -R appuser:appuser /app

# Copy binary and config with correct ownership
COPY --from=builder --chown=appuser:appuser /build/main .
COPY --from=builder --chown=appuser:appuser /build/pkg/config/config.docker.yaml /app/config/config.yaml

# Set permissions for execution and config read
RUN chmod 755 /app/main && \
    chmod -R 755 /app/config && \
    chmod -R 755 /app/certs

# Set environment variables
ENV CONFIG_FILE=/app/config/config.yaml \
    SERVER_PORT=8000 \
    SERVER_MODE=development \
    SERVER_TIMEOUT=30s \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=elhadi \
    DB_PASSWORD=test123 \
    DB_NAME=compass \
    DB_SSLMODE=disable \
    DB_TIMEZONE=UTC \
    DB_MAX_OPEN_CONNS=25 \
    DB_MAX_IDLE_CONNS=25 \
    DB_CONN_MAX_LIFETIME=1h \
    DB_POOL_SIZE=10 \
    DB_MIN_IDLE_CONNS=5 \
    DB_RETRY_ATTEMPTS=5 \
    DB_RETRY_DELAY=5s \
    REDIS_HOST=redis \
    REDIS_PORT=6380 \
    REDIS_PASSWORD="" \
    REDIS_DB=0 \
    JWT_SECRET="" \
    JWT_ISSUER=compass-api \
    JWT_EXPIRY_HOURS=24 \
    LOG_LEVEL=debug \
    LOG_FORMAT=json \
    TZ=UTC \
    USE_HTTPS=false

# Switch to non-root user
USER appuser

# Expose ports (both HTTP and HTTPS)
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Run the binary
CMD ["./main"] 