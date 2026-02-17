# syntax=docker/dockerfile:1

# ============================================
# Stage 1: Linting
# ============================================
FROM golangci/golangci-lint:v1.64-alpine AS linter

WORKDIR /app

COPY go.mod go.sum ./
COPY .golangci.yml ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN golangci-lint run --timeout=5m

# ============================================
# Stage 2: Testing
# ============================================
FROM golang:1.24-alpine AS tester

WORKDIR /app

# Install build dependencies including gcc for race detector
RUN apk add --no-cache git make gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run tests with race detector and coverage
RUN CGO_ENABLED=1 go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# ============================================
# Stage 3: Builder
# ============================================
FROM golang:1.24-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE
ARG BUILD_USER=docker

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build with version info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
    -X github.com/prometheus/common/version.Version=${VERSION} \
    -X github.com/prometheus/common/version.Revision=${COMMIT} \
    -X github.com/prometheus/common/version.Branch=main \
    -X github.com/prometheus/common/version.BuildUser=${BUILD_USER} \
    -X github.com/prometheus/common/version.BuildDate=${BUILD_DATE}" \
    -o eseries_exporter ./cmd/eseries_exporter

# ============================================
# Stage 4: Runtime
# ============================================
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /

# Copy binary from builder
COPY --from=builder /app/eseries_exporter /eseries_exporter

# Use nobody user for security
USER nobody

EXPOSE 9313

ENTRYPOINT ["/eseries_exporter"]
CMD ["--config.file=/eseries_exporter.yaml"]

# Labels
LABEL org.opencontainers.image.title="NetApp E-Series Exporter" \
      org.opencontainers.image.description="Prometheus exporter for NetApp E-Series storage systems" \
      org.opencontainers.image.vendor="sckyzo" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.source="https://github.com/sckyzo/eseries_exporter"
