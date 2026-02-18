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
ID: a1b2c3d4-e5f6-7890-abcd-ef1234567890	name: my-storage-system
```

### Querying Metrics

To query a storage system using the `default` module:
```
curl "http://localhost:9313/eseries?target=a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

To query a different target using the `status-only` module:
```
curl "http://localhost:9313/eseries?target=<storage-system-id>&module=status-only"
```

## Prometheus Configuration

The exporter uses **modules** to define different monitoring profiles. Modules are configured in `eseries_exporter.yaml` and called via the `module` URL parameter.

Available modules (from examples/eseries_exporter.yaml):
- `default` - All collectors enabled (full monitoring)
- `status-only` - Only status collectors (storage-systems, drives, hardware-inventory)
- `performance` - Performance metrics (controller-statistics, system-statistics, drive-statistics)
- `capacity` - Capacity metrics (storage-systems, storage-pools, volumes)

### Basic Configuration

Add the E-Series exporter to your `prometheus.yml`:

```yaml
scrape_configs:
  # Scrape exporter's own metrics
  - job_name: 'eseries-exporter'
    static_configs:
      - targets: ['localhost:9313']

  # Scrape E-Series storage systems with default module
  - job_name: 'eseries'
    scrape_interval: 60s
    scrape_timeout: 30s
    metrics_path: /eseries
    params:
      module: [default]  # Use the 'default' module from eseries_exporter.yaml
    static_configs:
      - targets:
          - a1b2c3d4-e5f6-7890-abcd-ef1234567890  # Storage system ID
          - b2c3d4e5-f6a7-8901-bcde-f12345678901  # Another storage system
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9313
```

**Note**: If `module` parameter is not specified, the `default` module is used automatically.

### Configuration with Multiple Modules

Use different modules for different monitoring needs. Each scrape job calls a specific module with `params: module: [module-name]`:

```yaml
scrape_configs:
  # Module 'default': Full monitoring (all collectors)
  # Corresponds to the 'default' module in eseries_exporter.yaml
  - job_name: 'eseries-full'
    scrape_interval: 60s
    scrape_timeout: 30s
    metrics_path: /eseries
    params:
      module: [default]  # Calls the 'default' module
    static_configs:
      - targets:
          - a1b2c3d4-e5f6-7890-abcd-ef1234567890
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9313

  # Module 'status-only': Status monitoring (faster, less data)
  # Corresponds to the 'status-only' module in eseries_exporter.yaml
  - job_name: 'eseries-status'
    scrape_interval: 30s
    scrape_timeout: 15s
    metrics_path: /eseries
    params:
      module: [status-only]  # Calls the 'status-only' module
    static_configs:
      - targets:
          - b2c3d4e5-f6a7-8901-bcde-f12345678901
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9313

  # Module 'capacity': Capacity monitoring only
  # Corresponds to the 'capacity' module in eseries_exporter.yaml
  - job_name: 'eseries-capacity'
    scrape_interval: 300s  # Every 5 minutes (capacity changes slowly)
    scrape_timeout: 30s
    metrics_path: /eseries
    params:
      module: [capacity]  # Calls the 'capacity' module
    static_configs:
      - targets:
          - a1b2c3d4-e5f6-7890-abcd-ef1234567890
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9313

  # Module 'performance': Performance metrics only
  # Corresponds to the 'performance' module in eseries_exporter.yaml
  - job_name: 'eseries-performance'
    scrape_interval: 30s
    scrape_timeout: 15s
    metrics_path: /eseries
    params:
      module: [performance]  # Calls the 'performance' module
    static_configs:
      - targets:
          - a1b2c3d4-e5f6-7890-abcd-ef1234567890
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9313
```

**How module selection works:**
1. Define modules in `eseries_exporter.yaml` with different collector sets
2. In Prometheus, use `params: module: [module-name]` to select which module to use
3. The URL becomes: `http://localhost:9313/eseries?target=xxx&module=module-name`
4. If no module is specified, `default` is used

### Configuration with Basic Authentication

If the exporter is protected with Basic Auth:

```yaml
scrape_configs:
  - job_name: 'eseries'
    scrape_interval: 60s
    scrape_timeout: 30s
    metrics_path: /eseries
    basic_auth:
      username: prometheus
      password: your-password
    static_configs:
      - targets:
          - a1b2c3d4-e5f6-7890-abcd-ef1234567890
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9313
```

### Configuration with TLS

For HTTPS endpoints:

```yaml
scrape_configs:
  - job_name: 'eseries'
    scrape_interval: 60s
    scrape_timeout: 30s
    metrics_path: /eseries
    scheme: https
    tls_config:
      ca_file: /etc/prometheus/certs/ca.crt
      cert_file: /etc/prometheus/certs/client.crt
      key_file: /etc/prometheus/certs/client.key
      # Or skip verification (not recommended for production)
      # insecure_skip_verify: true
    static_configs:
      - targets:
          - a1b2c3d4-e5f6-7890-abcd-ef1234567890
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: exporter.example.com:9313
```

### Service Discovery with File-based Configuration

For dynamic storage system discovery:

Create `eseries_targets.yml`:
```yaml
- targets:
    - a1b2c3d4-e5f6-7890-abcd-ef1234567890
    - b2c3d4e5-f6a7-8901-bcde-f12345678901
  labels:
    env: production
    site: datacenter1
```

Configure Prometheus:
```yaml
scrape_configs:
  - job_name: 'eseries'
    scrape_interval: 60s
    scrape_timeout: 30s
    metrics_path: /eseries
    file_sd_configs:
      - files:
          - /etc/prometheus/eseries_targets.yml
        refresh_interval: 5m
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9313
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
