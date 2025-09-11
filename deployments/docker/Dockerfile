# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -tags netgo -installsuffix netgo \
    -o bin/gotak-server \
    ./cmd/gotak-server

# Final stage
FROM scratch

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/bin/gotak-server /gotak-server

# Copy configuration
COPY --from=builder /app/config/server.yaml /config/server.yaml

# Expose ports
EXPOSE 8087 8089 8080

# Add metadata labels
LABEL org.opencontainers.image.title="GoTAK Server"
LABEL org.opencontainers.image.description="High-performance TAK server for military operations"
LABEL org.opencontainers.image.vendor="GoTAK Project"
LABEL org.opencontainers.image.licenses="MIT"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/gotak-server", "health"]

# Run as non-root user
USER 65534:65534

# Set the entrypoint
ENTRYPOINT ["/gotak-server"]

# Default command line arguments
CMD ["-config", "/config/server.yaml"]
