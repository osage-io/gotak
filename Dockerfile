# Multi-stage Dockerfile for GoTAK Production Deployment

# Web build stage
FROM node:20-alpine AS web-builder

# Set working directory
WORKDIR /web

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci

# Copy web source code
COPY web/ ./

# Build the web application
RUN npm run build

# Go build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make postgresql-client

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=1.0.0
ARG BUILD_TIME=dev
ARG GIT_COMMIT=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X 'main.version=${VERSION}' -X 'main.buildDate=${BUILD_TIME}' -X 'main.commit=${GIT_COMMIT}'" \
    -o gotak-server ./cmd/gotak-server

# Build client tools
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s" \
    -o gotak-client ./cmd/gotak-client

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    postgresql-client \
    netcat-openbsd \
    bash \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S gotak && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G gotak -g gotak gotak

# Set working directory
WORKDIR /app

# Copy certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binaries from builder stage
COPY --from=builder /build/gotak-server /app/
COPY --from=builder /build/gotak-client /app/

# Copy web build from web-builder stage
COPY --from=web-builder /web/dist /app/web

# Copy configuration files and scripts
COPY --chown=gotak:gotak config/ /app/config/
COPY --chown=gotak:gotak migrations/ /app/migrations/
COPY --chown=gotak:gotak scripts/entrypoint.sh /app/
COPY --chown=gotak:gotak scripts/healthcheck.sh /app/
COPY --chown=gotak:gotak scripts/migrate.sh /app/
COPY --chown=gotak:gotak scripts/backup.sh /app/
COPY --chown=gotak:gotak scripts/validate-config.sh /app/

# Make scripts executable
RUN chmod +x /app/entrypoint.sh /app/healthcheck.sh /app/migrate.sh /app/backup.sh /app/validate-config.sh

# Create necessary directories
RUN mkdir -p /app/logs /app/data /app/certs /app/web && \
    chown -R gotak:gotak /app/logs /app/data /app/certs /app/web

# Switch to non-root user
USER gotak

# Expose ports
EXPOSE 8080 8087 8089

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /app/healthcheck.sh

# Set environment variables
ENV GIN_MODE=release
ENV GOTAK_CONFIG_PATH=/app/config/production.yaml
ENV GOTAK_LOG_LEVEL=info
ENV GOTAK_DATA_DIR=/app/data
ENV GOTAK_LOG_DIR=/app/logs

# Use entrypoint script
ENTRYPOINT ["/app/entrypoint.sh"]

# Default command
CMD ["/app/gotak-server"]
