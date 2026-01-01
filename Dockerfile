# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o eseries_exporter ./cmd/eseries_exporter

# Final stage
FROM alpine:3.19

WORKDIR /

COPY --from=builder /app/eseries_exporter /eseries_exporter

EXPOSE 9313

USER nobody

ENTRYPOINT ["/eseries_exporter"]
CMD ["--config.file=/eseries_exporter.yaml"]