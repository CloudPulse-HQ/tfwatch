# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Backend auto-detection for Terraform Cloud and S3
- OpenTelemetry metric publishing via gRPC
- `--list` mode for printing dependencies without publishing
- `--phase` flag for tagging metrics with plan/apply phase
- Executive Overview Grafana dashboard with stats, charts, and version matrix
- Dependency Explorer Grafana dashboard with searchable tables
- Docker Compose observability stack (OTEL Collector, Prometheus, Grafana)
- Auto-provisioned Grafana dashboards
- 10 example Terraform repos for testing
- CI workflow with coverage reporting and PR validation
- Automatic semantic versioning and release on merge to main
- GitHub Pages website
- Dependabot for Go modules and GitHub Actions

### Changed
- Module path updated to `github.com/CloudPulse-HQ/tfwatch`
- `--dir` defaults to current directory (no longer required)

### Removed
- `--terraform-version` flag (always auto-detected now)
- `--workspace` and `--backend` flags (auto-detected from .tf files)
