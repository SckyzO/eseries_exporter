# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0] - 2026-01-01

### Major Changes (Breaking)
- **Architecture**: Move to standard Go project layout (`cmd/`, `internal/`).
- **Dependencies**: Update to Go 1.22+.
- **Logging**: Migrate from `go-kit/log` to `log/slog` (standard library).
- **Flags**: Migrate to `alecthomas/kingpin/v2`.
- **Config**: Enforce strict YAML parsing for configuration files. Fields unknown to the configuration schema will now cause an error.

### Improvements
- **CI/CD**: Add GitHub Actions workflows for robust linting (golangci-lint), testing (race detection), and multi-architecture builds.
- **Tests**: Fix flaky tests and improve test coverage. Add support for recent API responses (arrays instead of objects for statistics endpoints).
- **Build**: Modernize Makefile with `lint`, `test`, `build`, and `docker` targets. Cleanup old build artifacts (`promu`, `CircleCI`).

## [1.0.0] - 2020-11-13
- Initial release
