[![Build Status](https://github.com/sckyzo/eseries_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/sckyzo/eseries_exporter/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/sckyzo/eseries_exporter?include_prereleases&sort=semver)](https://github.com/sckyzo/eseries_exporter/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/sckyzo/eseries_exporter/total)
[![codecov](https://codecov.io/gh/sckyzo/eseries_exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/sckyzo/eseries_exporter)

# NetApp E-Series Prometheus exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/sckyzo/eseries_exporter)](https://goreportcard.com/report/github.com/sckyzo/eseries_exporter)
[![Go Reference](https://pkg.go.dev/badge/github.com/sckyzo/eseries_exporter.svg)](https://pkg.go.dev/github.com/sckyzo/eseries_exporter)

The E-Series exporter collects metrics from NetApp E-Series via the SANtricity Web Services Proxy.

This exporter is intended to query multiple E-Series controllers from an external host.

The `/eseries` metrics endpoint exposes E-Series metrics and requires the `target` parameter.

The `/metrics` endpoint exposes Go and process metrics for this exporter.

## 🚀 Features

- **Modern Go 1.24+** with latest dependencies
- **Structured logging** with `log/slog` (replace go-kit/log)
- **TLS support** with configurable certificates
- **BasicAuth protection** for all endpoints
- **Multi-architecture builds** (AMD64, ARM64, ARM)
- **Comprehensive CI/CD** with GitHub Actions
- **Security scanning** (Trivy, CodeQL, GoSec)
- **Modular architecture** with clean separation (cmd/, internal/)

## 📋 Collectors

Collectors are enabled or disabled via a config file.

| Name | Description | Default | Enabled |
|------|-------------|---------|---------|
| drives | Collect status information about drives | ✅ | Yes |
| drive-statistics | Collect statistics on drives | ⚠️ | No |
| controller-statistics | Collect controller statistics | ✅ | Yes |
| storage-systems | Collect status information about storage systems | ✅ | Yes |
| system-statistics | Collect storage system statistics | ✅ | Yes |
| hardware-inventory | Collect hardware inventory statuses | ✅ | Yes |
| **interface-statistics** | Collect network interface performance metrics | ✅ | Yes |
| **volume-statistics** | Collect storage volume performance metrics | ✅ | Yes |
| **system-health** | Collect system health and capacity metrics | ✅ | Yes |

## 🛡️ Security Features

### TLS Support

Enable TLS encryption for all web endpoints:

```bash
eseries_exporter \
  --web.tls-enabled \
  --web.tls-cert-file=/path/to/cert.pem \
  --web.tls-key-file=/path/to/key.pem \
  --web.listen-address=:9313
```

### Basic Authentication

Protect all endpoints with BasicAuth:

```bash
eseries_exporter \
  --web.auth-enabled \
  --web.auth.username=monitor \
  --web.auth.password-file=/etc/eseries_exporter/password \
  --web.listen-address=:9313
```

Or use a password file directly:

```bash
echo "secure-password" > /etc/eseries_exporter/password
chmod 600 /etc/eseries_exporter/password

eseries_exporter \
  --web.auth-enabled \
  --web.auth.username=monitor \
  --web.auth.password-file=/etc/eseries_exporter/password
```

### Prometheus Exporter Toolkit Format

For complex configurations, use the standard Prometheus Exporter Toolkit format:

```yaml
# web-config.yml
tls_server_config:
  cert_file: /path/to/cert.pem
  key_file: /path/to/key.pem

basic_auth_users:
  monitor: $2b$14$hashed_password_here
```

```bash
eseries_exporter --web.config.file=web-config.yml
```

## ⚙️ Configuration

The configuration defines targets that are to be queried. Example:

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

### Command Line Options

```
eseries_exporter --help
```

| Flag | Description | Default |
|------|-------------|---------|
| `--config.file` | Path to exporter config file | `eseries_exporter.yaml` |
| `--web.listen-address` | Address to listen on for web interface and telemetry | `:9313` |
| `--log.level` | Log level (debug, info, warn, error) | `info` |
| `--log.format` | Log format (text, json) | `text` |
| `--web.tls-enabled` | Enable TLS for web endpoint | `false` |
| `--web.tls-cert-file` | Path to TLS certificate file | - |
| `--web.tls-key-file` | Path to TLS private key file | - |
| `--web.auth-enabled` | Enable BasicAuth for web endpoint | `false` |
| `--web.auth.username` | Username for BasicAuth | `admin` |
| `--web.auth.password` | Password for BasicAuth | - |
| `--web.auth.password-file` | Path to password file for BasicAuth | - |
| `--web.config.file` | Path to web config file for TLS/BasicAuth | - |
| `--version` | Show application version | - |

### Usage Examples

This exporter could then be queried via one of these commands. The `eseries2` target will only run the `drives` and `storage-systems` collectors.

**Basic HTTP (unauthenticated):**
```bash
curl http://localhost:9313/eseries?target=eseries1
curl http://localhost:9313/eseries?target=eseries2&module=status-only
```

**HTTPS with TLS:**
```bash
curl -k https://localhost:9313/eseries?target=eseries1
```

**BasicAuth with curl:**
```bash
curl -u monitor:secure-password https://localhost:9313/eseries?target=eseries1
```

If no `timeout` is defined the default is `10`.

If the HTTP schema used for `proxy_url` is `https` then the exporter will attempt to use the system CA truststore as well as any root CA specified with `root_ca` option. By default certificate verification is enabled, set `insecure_ssl` to disable SSL verification.

## 🔗 Dependencies

This exporter expects to communicate with SANtricity Web Services Proxy API and that your storage controllers are already setup to be accessed through that API.

This repo provides a Docker based approach to running the Web Services Proxy:

```bash
cd webservices_proxy
docker build -t webservices_proxy .
```

The above Docker container will have `admin:admin` as the credentials and can be run using a command like the following:

```bash
docker run -d --rm -it --name webservices_proxy --network host -e ACCEPT_EULA=true \
-v /var/lib/eseries_webservices_proxy/working:/opt/netapp/webservices_proxy/working \
webservices_proxy
```

**NOTE**: During testing it seemed in order for a Docker based proxy to communicate with E-Series controllers the container had to use the host's network.

Example of setting up the Web Services Proxy with an E-Series system. Replace `PASSWORD` with the password for your E-Series system. Replace `ID` with the name of your system. With `IP1` and `IP2` with IP addresses of your controllers for the system.

```bash
curl -X POST -u admin:admin "http://localhost:8080/devmgr/v2/storage-systems" \
-H  "accept: application/json" -H  "Content-Type: application/json" \
-d '{
  "id": "ID", "controllerAddresses": [ "IP1" , "IP2" ],
  "acceptCertificate": true, "validate": false, "password": "PASSWORD"
}'
```

If your storage systems have UUID style IDs this is a way to query the names for each ID:

```bash
$ curl -u admin:admin http://localhost:8080/devmgr/v2/storage-systems 2>/dev/null | \
  jq -r '.[] | "ID: \(.id)\tname: \(.name)"'
ID: f0d2fadc-3e16-46c5-b62e-c9ab6d430b50    name: eseries1
ID: 25db8d36-6732-495d-b693-8add202750d6    name: eseries2
```

## 🐳 Docker

### Official Images

Images are available on Docker Hub:

```bash
docker run -d -p 9313:9313 \
  -v "$(pwd)/eseries_exporter.yaml:/eseries_exporter.yaml:ro" \
  sckyzo/eseries_exporter:latest
```

### Docker Compose

This repo provides a Docker Compose file that can be used to run both the Web Services Proxy and this exporter.

```bash
docker-compose up -d
```

See [dependencies section](#dependencies) for steps necessary to bootstrap the Web Services Proxy.

### Docker with TLS

```bash
docker run -d -p 443:9313 \
  -v "$(pwd)/eseries_exporter.yaml:/eseries_exporter.yaml:ro" \
  -v "$(pwd)/certs:/certs:ro" \
  sckyzo/eseries_exporter:latest \
  --web.tls-enabled \
  --web.tls-cert-file=/certs/cert.pem \
  --web.tls-key-file=/certs/key.pem
```

### Docker with BasicAuth

```bash
docker run -d -p 9313:9313 \
  -v "$(pwd)/eseries_exporter.yaml:/eseries_exporter.yaml:ro" \
  -v "$(pwd)/password:/password:ro" \
  sckyzo/eseries_exporter:latest \
  --web.auth-enabled \
  --web.auth.username=monitor \
  --web.auth.password-file=/password
```

## 📦 Installation

### Binary Release

Download the [latest release](https://github.com/sckyzo/eseries_exporter/releases)

Add the user that will run `eseries_exporter`

```bash
groupadd -r eseries_exporter
useradd -r -d /var/lib/eseries_exporter -s /sbin/nologin -M -g eseries_exporter eseries_exporter
```

Install compiled binaries after extracting tar.gz from release page.

```bash
cp /tmp/eseries_exporter /usr/local/bin/eseries_exporter
```

### Package Managers

#### Homebrew (macOS/Linux)

```bash
brew tap sckyzo/eseries_exporter
brew install eseries_exporter
```

#### APT (Debian/Ubuntu)

```bash
wget -qO- https://apt.sckyzo.com/key.gpg | sudo apt-key add -
echo "deb https://apt.sckyzo.com/ eseries_exporter main" | sudo tee /etc/apt/sources.list.d/eseries_exporter.list
sudo apt-get update && sudo apt-get install eseries_exporter
```

#### YUM (RHEL/CentOS)

```bash
sudo yum install -y https://rpm.sckyzo.com/eseries_exporter_latest_amd64.rpm
```

### Docker

See [Docker section](#docker) for containerized installation.

## 🏗️ Build from source

### Prerequisites

- Go 1.24 or later
- Git

### Build

To produce the `eseries_exporter` binary:

```bash
git clone https://github.com/sckyzo/eseries_exporter.git
cd eseries_exporter
go build ./cmd/eseries_exporter
```

### Make targets

```bash
make build        # Build the binary
make test         # Run tests
make lint         # Run golangci-lint
make clean        # Clean build artifacts
make docker       # Build Docker image
make release      # Create release with GoReleaser
```

### Development

```bash
# Run tests with coverage
make test

# Run linting
make lint

# Build for multiple platforms
make build-all

# Run in development mode
go run ./cmd/eseries_exporter --help
```

## 📊 Prometheus configurations

The following example assumes this exporter is running on the Prometheus server and communicating to the remote E-Series API.

```yaml
- job_name: eseries
  metrics_path: /eseries
  static_configs:
  - targets:
    - eseries1
    - eseries2
  relabel_configs:
  - source_labels: [__address__]
    target_label: __param_target
  - source_labels: [__param_target]
    target_label: instance
  - target_label: __address__
    replacement: 127.0.0.1:9313
- job_name: eseries-metrics
  metrics_path: /metrics
  static_configs:
  - targets:
    - localhost:9313
```

### Secure Prometheus Configuration

If using TLS and BasicAuth:

```yaml
- job_name: eseries-secure
  metrics_path: /eseries
  basic_auth:
    username: monitor
    password: secure-password
  scheme: https
  tls_config:
    insecure_skip_verify: false
    ca_file: /etc/prometheus/ca.pem
    cert_file: /etc/prometheus/cert.pem
    key_file: /etc/prometheus/key.pem
  static_configs:
  - targets:
    - eseries1
    - eseries2
  relabel_configs:
  - source_labels: [__address__]
    target_label: __param_target
  - source_labels: [__param_target]
    target_label: instance
  - target_label: __address__
    replacement: exporter.example.com:443
```

The following is an example if your E-Series web proxy is using UUIDs or other cryptic IDs:

```yaml
- job_name: eseries
  metrics_path: /eseries
  static_configs:
  - targets:
    - f0d2fadc-3e16-46c5-b62e-c9ab6d430b50
    labels:
      name: eseries1
  - targets:
    - 25db8d36-6732-495d-b693-8add202750d6
    labels:
      name: eseries2
  relabel_configs:
  - source_labels: [__address__]
    target_label: __param_target
  - target_label: __address__
    replacement: 127.0.0.1:9313
  - source_labels: [name]
    target_label: instance
  - regex: '^(name)$'
    action: labeldrop
```

## 🔄 CI/CD

This project uses GitHub Actions for CI/CD:

- **Continuous Integration**: Runs on every push and PR
  - Code quality checks with `golangci-lint`
  - Unit tests with coverage reporting
  - Multi-OS builds (Linux, Windows, macOS)
  - Security scanning with CodeQL and Trivy

- **Continuous Deployment**: Runs on git tags
  - Automated releases with GoReleaser
  - Multi-platform binary generation
  - Docker image builds and pushes
  - Changelog generation

- **Security**: Continuous monitoring
  - Dependency vulnerability scanning
  - Static code analysis
  - Supply chain security

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with race detector
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.txt ./...

# Run specific test
go test -v ./cmd/eseries_exporter/...

# Run integration tests
make test-integration
```

## 📈 Monitoring

The exporter exposes several metrics for monitoring:

- **Exporter metrics** (`/metrics`):
  - `eseries_exporter_collector_duration_seconds`
  - `eseries_exporter_collect_error`

- **Target metrics** (`/eseries`):
  - All E-Series storage system metrics
  - Controller statistics
  - Drive status and statistics
  - Hardware inventory

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run tests and linting: `make test lint`
5. Commit your changes: `git commit -am 'Add feature'`
6. Push to the branch: `git push origin feature-name`
7. Create a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- NetApp for the E-Series SANtricity Web Services Proxy API
- The Prometheus community for exporter guidelines
- Contributors and maintainers of Go dependencies

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/sckyzo/eseries_exporter/issues)
- **Discussions**: [GitHub Discussions](https://github.com/sckyzo/eseries_exporter/discussions)
- **Security**: Please report security vulnerabilities privately

## 📚 API Documentation

This exporter uses the official NetApp E-Series Web Services Proxy API. For detailed API information, see:

- [NetApp E-Series Web Services Proxy Documentation](web_services_proxy.pdf)
- [SANtricity Web Services API Reference](https://github.com/NetApp/eseries-webservices)