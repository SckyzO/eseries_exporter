# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Refactor
- **Architecture**: Move to standard Go project layout (`cmd/`, `internal/`).
- **Dependencies**: Update to Go 1.22+.
- **Logging**: Migrate from `go-kit/log` to `log/slog` (standard library).
- **Flags**: Migrate to `alecthomas/kingpin/v2`.
- **Config**: Add strict YAML parsing for better error detection.
- **Tests**: Fix flaky tests and improve test coverage. Add support for recent API responses (arrays instead of objects for statistics).

### Added
- **CI/CD**: Add GitHub Actions workflows for linting, testing, and multi-arch build.
- **Makefile**: Modernize Makefile with lint, test, and build targets.

## [1.0.0] - 2020-11-13
- Initial release