## 2.0.0 / 2025-11-21

### Major Changes and Refactoring

This release represents a complete modernization of the eseries_exporter with significant architectural improvements and new security features.

#### 🚀 New Features

- **Modern Architecture**: Complete project restructure with cmd/, internal/ package organization
- **Go 1.24+ Support**: Upgraded from Go 1.17 to Go 1.24.10 with latest toolchain
- **40+ New Performance Metrics**: Added comprehensive monitoring for network interfaces, storage volumes, and system health
  - **Interface Statistics**: 9 new metrics (IOPS, throughput, response time, link utilization)
  - **Volume Statistics**: 11 new metrics (IOPS, throughput, cache performance, latency)
  - **System Health**: 20 new metrics (firmware, capacity, uptime, security status)
- **TLS Support**: Added comprehensive TLS encryption for all web endpoints
  - `--web.tls-enabled` flag for enabling TLS
  - `--web.tls-cert-file` and `--web.tls-key-file` for certificate configuration
  - TLS 1.2+ minimum version for security
- **BasicAuth Protection**: Added Basic Authentication for all endpoints
  - `--web.auth-enabled` flag for enabling authentication
  - `--web.auth.username` and `--web.auth.password` for credentials
  - `--web.auth.password-file` for secure password storage
  - Default username: `admin` if not specified
- **Prometheus Exporter Toolkit Support**: Added `--web.config.file` for standard configuration format
- **Structured Logging**: Migrated from go-kit/log to Go's built-in log/slog
  - Support for both text and JSON log formats
  - Configurable log levels (debug, info, warn, error)
  - Consistent structured logging throughout the application

#### 🔧 Improvements

- **CI/CD Modernization**: 
  - Replaced CircleCI with GitHub Actions
  - Comprehensive CI workflow with linting, testing, and builds
  - Automated release pipeline with GoReleaser
  - Multi-OS testing (Ubuntu, Windows, macOS)
  - Security scanning with Trivy, CodeQL, and GoSec
- **Modern Dependencies**: Updated all Go modules to latest stable versions
  - github.com/prometheus/client_golang v1.23.2
  - github.com/prometheus/common v0.67.3
  - github.com/alecthomas/kingpin/v2 v2.4.0
- **Module Path Migration**: Changed from github.com/treydock/eseries_exporter to github.com/sckyzo/eseries_exporter
- **Docker Improvements**: Multi-architecture image builds (AMD64, ARM64, ARM)
- **Documentation**: Complete README overhaul with:
  - Security configuration examples
  - Docker usage with TLS/BasicAuth
  - Package manager installation guides
  - Comprehensive CLI documentation
  - Prometheus configuration examples

#### 📊 Enhanced Metrics Coverage

- **60+ Total Metrics Available**: Dramatically increased monitoring capabilities
- **Network Interface Monitoring**: IOPS, throughput, latency, and error tracking
- **Storage Volume Analytics**: Performance metrics with cache hit ratios
- **System Health Dashboard**: Firmware versions, capacity planning, uptime monitoring
- **Proactive Monitoring**: Early detection of performance issues and capacity constraints

#### �️ Security Enhancements

- **TLS Encryption**: All web endpoints can now be secured with TLS
- **Authentication**: BasicAuth protection for metrics endpoints
- **Security Scanning**: Automated vulnerability scanning in CI/CD
- **Modern Standards**: Compliance with Prometheus Exporter Toolkit guidelines

#### 🏗️ Breaking Changes

- **Module Path**: Import path changed from `github.com/treydock/eseries_exporter` to `github.com/sckyzo/eseries_exporter`
- **New Required Dependencies**: 
  - github.com/alecthomas/kingpin/v2 (replacing gopkg.in/alecthomas/kingpin.v2)
  - Standard library log/slog (replacing github.com/go-kit/log)

#### 🐳 Docker Changes

- **New Docker Hub Organization**: Images now published as `sckyzo/eseries_exporter`
- **Multi-Architecture**: Support for ARM64 and ARM architectures
- **Security Features**: Docker images support TLS and BasicAuth configurations

#### 📋 Development Improvements

- **Modern Build System**: GoReleaser for automated releases
- **Code Quality**: golangci-lint integration with optimized rules
- **Test Coverage**: Enhanced testing with race detection and coverage reporting
- **Documentation**: Comprehensive README with security examples

#### 🛡️ Security Enhancements

- **TLS Encryption**: All web endpoints can now be secured with TLS
- **Authentication**: BasicAuth protection for metrics endpoints
- **Security Scanning**: Automated vulnerability scanning in CI/CD
- **Modern Standards**: Compliance with Prometheus Exporter Toolkit guidelines

#### 🏗️ Breaking Changes

- **Module Path**: Import path changed from `github.com/treydock/eseries_exporter` to `github.com/sckyzo/eseries_exporter`
- **New Required Dependencies**: 
  - github.com/alecthomas/kingpin/v2 (replacing gopkg.in/alecthomas/kingpin.v2)
  - Standard library log/slog (replacing github.com/go-kit/log)

#### 🐳 Docker Changes

- **New Docker Hub Organization**: Images now published as `sckyzo/eseries_exporter`
- **Multi-Architecture**: Support for ARM64 and ARM architectures
- **Security Features**: Docker images support TLS and BasicAuth configurations

#### 📋 Development Improvements

- **Modern Build System**: GoReleaser for automated releases
- **Code Quality**: golangci-lint integration with optimized rules
- **Test Coverage**: Enhanced testing with race detection and coverage reporting
- **Documentation**: Comprehensive README with security examples

### Migration Guide

#### For Users

1. **Update Import Paths**: Change all imports from `github.com/treydock/eseries_exporter` to `github.com/sckyzo/eseries_exporter`
2. **New Security Features**: Consider enabling TLS and BasicAuth for production deployments
3. **Docker**: Update image references from `treydock/eseries_exporter` to `sckyzo/eseries_exporter`

#### For Developers

1. **Dependencies**: Update go.mod with new module path and latest dependencies
2. **Logging**: Migrate from go-kit/log to log/slog (LoggerAdapter provided for compatibility)
3. **Build**: Use new Makefile targets and GoReleaser for releases

### Compatibility

- **API Unchanged**: All existing E-Series metrics and collectors remain compatible
- **Configuration**: Existing YAML configurations continue to work without changes
- **Prometheus**: No changes required for existing Prometheus configurations

### Known Issues

- None reported

### Contributors

- SckyzO - Complete refactoring and modernization

---

## 1.3.0 / 2022-03-08

* Improved SSL support for communicating with proxy API
* Update to Go 1.17
* Update Go module dependencies

## 1.3.0-rc.0 / 2021-08-24

* Improved SSL support for communicating with proxy API

## 1.2.1 / 2021-07-09

* Avoid errors if duplicate drives are encountered

## 1.2.0 / 2021-04-23

### Changes

* Update to Go 1.16
* Update Go module dependencies

## 1.1.0 / 2021-03-19

### Changes

* Add hardware-inventory collector

## 1.0.0 / 2020-11-25

### BREAKING CHANGES

* Remove --exporter.use-cache flag and all caching logic
* Remove ID related labels from all metrics as this will be instance label
* Refactor drive-statistics collector
* Refactor system-statistics collector
* CPU utilization metrics are now ratios of 0.0-1.0, add _ratio suffix to metrics
* Remove metrics
  * eseries_drive_combined_iops, eseries_drive_combined_throughput_bytes_per_second
  * eseries_drive_read_ops, eseries_drive_read_iops, eseries_drive_read_throughput_mb_per_second
  * eseries_drive_write_ops, eseries_drive_write_iops, eseries_drive_write_throughput_mb_per_second
  * eseries_system_read_ops, eseries_system_read_iops, eseries_system_read_throughput_mb_per_second
  * eseries_system_write_ops, eseries_system_write_iops, eseries_system_write_throughput_mb_per_second
  * eseries_system_cache_hit_bytes_percent, eseries_system_combined_iops
  * eseries_system_combined_throughput_mb_per_second, eseries_system_ddp_bytes_percent
  * eseries_system_full_stripe_writes_bytes_percent, eseries_system_random_ios_percent
* Rename metrics to change units of measurement
  * eseries_drive_combined_response_time_milliseconds to eseries_drive_combined_response_time_seconds
  * eseries_drive_read_response_time_milliseconds to eseries_drive_read_response_time_seconds
  * eseries_drive_write_response_time_milliseconds to eseries_drive_write_response_time_seconds
  * eseries_drive_combined_throughput_mb_per_second to eseries_drive_combined_throughput_bytes_per_second
  * eseries_system_combined_hit_response_time_milliseconds to eseries_system_combined_hit_response_time_seconds
  * eseries_system_combined_response_time_milliseconds to eseries_system_combined_response_time_seconds
  * eseries_system_read_hit_response_time_milliseconds to eseries_system_read_hit_response_time_seconds
  * eseries_system_read_response_time_milliseconds to eseries_system_read_response_time_seconds
  * eseries_system_write_hit_response_time_milliseconds to eseries_system_write_hit_response_time_seconds
  * eseries_system_write_response_time_milliseconds to eseries_system_write_response_time_seconds
* Add metrics
  * eseries_drive_idle_time_seconds_total, eseries_drive_other_ops_total, eseries_drive_other_time_seconds_total
  * eseries_drive_read_bytes_total, eseries_drive_read_ops_total, eseries_drive_read_time_seconds_total
  * eseries_drive_recovered_errors_total, eseries_drive_retried_ios_total, eseries_drive_timeouts_total
  * eseries_drive_unrecovered_errors_total
  * eseries_drive_write_bytes_total, eseries_drive_write_ops_total, eseries_drive_write_time_seconds_total
  * eseries_drive_queue_depth_total, eseries_drive_random_ios_total, eseries_drive_random_bytes_total

### Improvements

* Update to Go 1.15 and update all dependencies
* Improve status metrics to always have all possible statuses and set 1 for current status
* Add controller-statistics collector
* Add webservices_proxy Docker container
* Add Docker Compose example

## 0.1.1 / 2020-04-03

* Minor fix to Docker container

## 0.1.0 / 2020-04-03

* Disable drive-statistics collector by default

## 0.0.1 / 2020-04-02

* Initial Release