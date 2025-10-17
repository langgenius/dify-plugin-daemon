FROM golang:1.23-alpine AS builder

ARG VERSION=unknown

# Install ca-certificates and timezone data for final stage
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy project files
COPY . .

# Build with optimizations and security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "\
    -s -w \
    -X 'github.com/langgenius/dify-plugin-daemon/internal/manifest.VersionX=${VERSION}' \
    -X 'github.com/langgenius/dify-plugin-daemon/internal/manifest.BuildTimeX=$(date -u +%Y-%m-%dT%H:%M:%S%z)'" \
    -o /app/main cmd/server/main.go

# Use Alpine for better permission handling with mounted volumes
FROM alpine:3.20

# Install ca-certificates for SSL/TLS
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user with specific UID/GID for consistency
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

# Set working directory
WORKDIR /app

# Create storage directory with proper permissions
RUN mkdir -p /app/api/storage && \
    chown -R appuser:appgroup /app && \
    chmod -R 755 /app

# Build args and environment
ARG PLATFORM=serverless
ENV PLATFORM=$PLATFORM
ENV GIN_MODE=release
ENV TZ=UTC

# Copy binary with proper ownership
COPY --from=builder --chown=appuser:appgroup /app/main /app/main

# Run as non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./main", "health"] || exit 1

# Run the server
ENTRYPOINT ["/app/main"]
