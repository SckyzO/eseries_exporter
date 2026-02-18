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
| **volumes** | Collect volume metrics (capacity, status, thin provisioning, mappings) | **Disabled** |
| **storage-pools** | Collect storage pool metrics (capacity, utilization, RAID status) | **Disabled** |

## Security (TLS & Basic Authentication)

The exporter supports TLS and Basic Authentication via the Prometheus exporter-toolkit. Create a web configuration file and pass it with `--web.config.file`:

```bash
eseries_exporter --web.config.file=web-config.yaml
```

### Basic Authentication Example

Generate a bcrypt password hash:
```bash
htpasswd -nBC 10 "" | tr -d ':\n'
```

Create `web-config.yaml`:
```yaml
basic_auth_users:
  prometheus: $2y$10$... # bcrypt hash
```

### TLS Example

```yaml
tls_server_config:
  cert_file: /path/to/server.crt
  key_file: /path/to/server.key
  min_version: TLS12
  max_version: TLS13
```

See `examples/web-config.yaml` and `examples/web-config-basic-auth.yaml` for complete examples.

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

## Installation & Usage

### 1. From Binaries (Systemd)

Download the latest release from the [Releases page](https://github.com/sckyzo/eseries_exporter/releases).

1.  **Create user**:
    ```bash
    useradd -r -s /sbin/nologin eseries_exporter
    ```

2.  **Install binary**:
    ```bash
    cp eseries_exporter /usr/local/bin/
    chmod +x /usr/local/bin/eseries_exporter
    ```

3.  **Configure**:
    Copy your `eseries_exporter.yaml` to `/etc/eseries_exporter.yaml`.

4.  **Install Service**:
    ```bash
    cp systemd/eseries_exporter.service /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable --now eseries_exporter
    ```

### 2. Docker

```bash
docker run -d -p 9313:9313 -v "$(pwd)/eseries_exporter.yaml:/eseries_exporter.yaml:ro" sckyzo/eseries_exporter
```

### Getting Storage System IDs

To retrieve the IDs of your E-Series storage systems from the Web Services Proxy:

```bash
curl -k -u admin https://localhost:8443/devmgr/v2/storage-systems | jq -r '.[] | "ID: \(.id)\tname: \(.name)"'
```

This will output:
```
ID: a2b84075-cfb6-468f-a53b-ea680868161a	name: da-nfs
```

### Querying Metrics

To query the `a2b84075-cfb6-468f-a53b-ea680868161a` target using the `default` module:
```
curl "http://localhost:9313/eseries?target=a2b84075-cfb6-468f-a53b-ea680868161a"
```

To query a different target using the `status-only` module:
```
curl "http://localhost:9313/eseries?target=<storage-system-id>&module=status-only"
```

## Development

### Requirements
- Go 1.24+
- Make
- Docker (for multi-stage builds)

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
