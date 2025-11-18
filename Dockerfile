# Multi-stage build for security and size optimization
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Create appuser for running the application
RUN adduser -D -g '' appuser

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/useragent-demo \
    ./cmd/demo

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies and security updates
RUN apk --no-cache add ca-certificates tzdata && \
    apk --no-cache upgrade

# Create non-root user
RUN adduser -D -g '' appuser

# Create directory for database with proper permissions
RUN mkdir -p /data && chown appuser:appuser /data

# Copy timezone data and certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /app/useragent-demo /app/useragent-demo

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /app

# Environment variables
ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    DB_PATH=/data/useragent.db \
    APP_ENV=production \
    LOG_LEVEL=info

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Volume for persistent data
VOLUME ["/data"]

# Run the application
ENTRYPOINT ["/app/useragent-demo"]
