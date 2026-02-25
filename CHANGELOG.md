# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.1](https://github.com/CloudPulse-HQ/tfwatch/compare/v0.1.0...v0.1.1) (2026-02-25)


### Bug Fixes

* add base href and favicon for GitHub Pages subdirectory ([cd0fb4a](https://github.com/CloudPulse-HQ/tfwatch/commit/cd0fb4a7b770ba2851f01ace8a194b530f73bfdb))
* add base href and favicon for GitHub Pages subdirectory ([45f19e0](https://github.com/CloudPulse-HQ/tfwatch/commit/45f19e0d08d0ef66b66f99ddd7a3efed4cb35ad6))

## [0.1.0](https://github.com/CloudPulse-HQ/tfwatch/compare/v0.0.1...v0.1.0) (2026-02-24)


### Features

* Go open source best practices ([71f87e5](https://github.com/CloudPulse-HQ/tfwatch/commit/71f87e5bab07003cad66ab99f66f111d11139a84))


### Bug Fixes

* add doc comments to exported types and functions ([92df1d0](https://github.com/CloudPulse-HQ/tfwatch/commit/92df1d03db54d14a1ad3577cd18beac1c90915b5))
* CI coverage threshold and Pages enablement ([9e26b4f](https://github.com/CloudPulse-HQ/tfwatch/commit/9e26b4f0baf167f5ffe96374302cad73c9778ea6))
* CI coverage threshold and Pages enablement ([9e26b4f](https://github.com/CloudPulse-HQ/tfwatch/commit/9e26b4f0baf167f5ffe96374302cad73c9778ea6))
* CI coverage threshold and Pages enablement ([895438e](https://github.com/CloudPulse-HQ/tfwatch/commit/895438e5da34d34d508aff94656c99ad3b795daa))
* **ci:** allow manual website deploy and fix paths trigger ([a00d5f4](https://github.com/CloudPulse-HQ/tfwatch/commit/a00d5f40c3b0aabe91b9ae0b561e0b232c0ce4ca))
* **ci:** allow manual website deploy and fix paths trigger ([a00d5f4](https://github.com/CloudPulse-HQ/tfwatch/commit/a00d5f40c3b0aabe91b9ae0b561e0b232c0ce4ca))
* **ci:** allow manual website deploy and fix paths trigger ([17aabd1](https://github.com/CloudPulse-HQ/tfwatch/commit/17aabd1244f635da3b54db2c266b235b0aa70c5b))
* **ci:** pin golangci-lint to v2 and skip merge commits in lint ([9b84c2a](https://github.com/CloudPulse-HQ/tfwatch/commit/9b84c2a90119e3ed11cf8082f7c99721236db613))
* **ci:** upgrade golangci-lint-action to v7 for v2 support ([28c4df6](https://github.com/CloudPulse-HQ/tfwatch/commit/28c4df66f83ae55d55ed2484fa7a6d6ec8d025f0))

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
