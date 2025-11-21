# Multi-stage Dockerfile for eseries_exporter
# Stage 1: Build stage
FROM --platform=$BUILDPLATFORM golang:1.24.10-alpine AS builder

# Install ca-certificates for HTTPS support
RUN apk add --no-cache ca-certificates git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# Support for custom version via --build-arg VERSION=2.1.0
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH

# Build with version information
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-s -w \
    -X github.com/prometheus/common/version.Version=${VERSION} \
    -X github.com/prometheus/common/version.Revision=$(git rev-parse --short HEAD 2>/dev/null || echo unknown) \
    -X github.com/prometheus/common/version.Branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown) \
    -X github.com/prometheus/common/version.BuildUser=$(whoami 2>/dev/null || echo unknown) \
    -X github.com/prometheus/common/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/eseries_exporter ./cmd/eseries_exporter

# Stage 2: Production stage
FROM --platform=$TARGETPLATFORM alpine:3.19

# Install ca-certificates for HTTPS and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 -S eseries && \
    adduser -D -u 1000 -S eseries -G eseries

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder --chown=eseries:eseries /app/eseries_exporter /app/

# Copy default configuration (if exists)
# COPY --chown=eseries:eseries config/eseries_exporter.yaml /app/config.yaml

# Set ownership and permissions
RUN chmod 755 /app/eseries_exporter

# Switch to non-root user
USER eseries

# Expose metrics port
EXPOSE 9313

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9313/health || exit 1

# Default configuration
ENV ESERIES_EXPORTER_CONFIG_FILE=/app/config.yaml
ENV ESERIES_EXPORTER_LOG_LEVEL=info
ENV ESERIES_EXPORTER_LOG_FORMAT=text

# Labels for better discoverability
LABEL maintainer="sckyzo <contact@sckyzo.com>" \
      org.opencontainers.image.title="eseries_exporter" \
      org.opencontainers.image.description="Prometheus exporter for NetApp E-Series storage systems" \
      org.opencontainers.image.url="https://github.com/sckyzo/eseries_exporter" \
      org.opencontainers.image.source="https://github.com/sckyzo/eseries_exporter" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
      org.opencontainers.image.created="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
      org.opencontainers.image.documentation="https://github.com/sckyzo/eseries_exporter/blob/main/README.md"

# Expose additional helpful information
LABEL org.opencontainers.image.vendor="SckyzO" \
      org.opencontainers.image.name="eseries-exporter" \
      org.opencontainers.image.os="linux" \
      org.opencontainers.image.architecture="amd64"

# Entry point with configurable arguments
ENTRYPOINT ["/app/eseries_exporter"]
CMD ["--help"]

# Alternative: For production, you might want to use:
# ENTRYPOINT ["/app/eseries_exporter", "--config.file=/app/config.yaml"]
