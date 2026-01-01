# NetApp E-Series Prometheus Exporter

[![Build Status](https://github.com/sckyzo/eseries_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/sckyzo/eseries_exporter/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/sckyzo/eseries_exporter)](https://goreportcard.com/report/github.com/sckyzo/eseries_exporter)

The E-Series exporter collects metrics from NetApp E-Series via the SANtricity Web Services Proxy (REST API).

This exporter is designed to query multiple E-Series controllers from an external host.

- The `/eseries` metrics endpoint exposes E-Series metrics and requires the `target` parameter.
- The `/metrics` endpoint exposes internal process metrics for this exporter.

## Architecture

This project follows standard Go layout:
- `cmd/eseries_exporter`: Main entry point.
- `internal/collectors`: Metric collectors logic.
- `internal/config`: Configuration handling.

## Collectors

Collectors are enabled or disabled via the config file.

| Name | Description | Default |
|------|-------------|---------|
| drives | Collect status information about drives | Enabled |
| drive-statistics | Collect statistics on drives | Disabled |
| controller-statistics | Collect controller statistics | Enabled |
| storage-systems | Collect status information about storage systems | Enabled |
| system-statistics | Collect storage system statistics | Enabled |
| hardware-inventory | Collect hardware inventory statuses | Enabled |

## Configuration

The configuration defines targets (modules) that are to be queried. Example `eseries_exporter.yaml`:

```yaml
modules:
  default:
    user: monitor
    password: secret
    proxy_url: http://localhost:8080
    timeout: 10
  status-only:
    user: monitor
    password: secret
    proxy_url: http://localhost:8080
    timeout: 10
    collectors:
      - drives
      - storage-systems
  ssl:
    user: monitor
    password: secret
    proxy_url: https://proxy.example.com
    root_ca: /etc/pki/tls/root.pem
    insecure_ssl: false
    timeout: 10
```

## Usage

### Binaries

Download the latest release for your platform.

```bash
./eseries_exporter --config.file=eseries_exporter.yaml
```

### Docker

```bash
docker run -d -p 9313:9313 -v "$(pwd)/eseries_exporter.yaml:/eseries_exporter.yaml:ro" sckyzo/eseries_exporter
```

### Querying Metrics

To query the `eseries1` target using the `default` module:
```
curl "http://localhost:9313/eseries?target=eseries1"
```

To query the `eseries2` target using the `status-only` module:
```
curl "http://localhost:9313/eseries?target=eseries2&module=status-only"
```

## Development

### Requirements
- Go 1.22+
- Make

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

## Dependencies

This exporter expects to communicate with SANtricity Web Services Proxy API. Ensure your storage controllers are accessible through that API.

See `webservices_proxy/` directory for a Docker-based approach to running the Web Services Proxy.