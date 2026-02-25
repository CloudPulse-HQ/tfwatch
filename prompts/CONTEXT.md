# tfwatch — Project Context

## Mission

tfwatch is a single-binary, zero-config CLI that reads Terraform lock files and publishes dependency metadata as OpenTelemetry metrics. It lets teams query module and provider versions across repos without grepping through codebases.

## Architecture

Single Go package (`github.com/CloudPulse-HQ/tfwatch`) with three core files:

| File | Responsibility |
|------|---------------|
| `parser.go` | Parse `.tf`, `.terraform.lock.hcl`, `modules.json`; auto-detect backends (Cloud, S3); auto-init |
| `collector.go` | Build and publish `terraform_dependency_version` Int64Gauge via OpenTelemetry gRPC |
| `main.go` | CLI entrypoint, flag parsing, OTEL SDK init, `--list` mode |
| `doc.go` | Package-level godoc |

Metric label schema: `backend_type`, `backend_org`, `backend_workspace`, `phase`, `type`, `dependency_name`, `dependency_source`, `dependency_version`, `terraform_version`.

## Commit Rules

- Conventional commits: `type(scope): description` (e.g., `fix(parser): handle empty lock files`)
- Valid types: `feat`, `fix`, `docs`, `chore`, `ci`, `test`, `refactor`, `style`
- Atomic commits — one logical change per commit
- Never amend published commits

## Branch Rules

- Always branch from latest `main`
- Naming: `<type>/<short-description>` (e.g., `feat/s3-backend`, `fix/parser-nil-check`)
- PRs target `main`; squash-merge preferred

## Go Conventions

- Single package, no internal/ or cmd/ hierarchy
- Table-driven tests with descriptive subtest names
- Minimum 80% test coverage (enforced in CI)
- golangci-lint v2: errcheck, govet, staticcheck, unused, ineffassign, misspell, revive, errorlint
- Wrap errors with `fmt.Errorf("context: %w", err)`
- Functions under 50 lines; extract helpers when needed

## Makefile Targets

`build`, `install`, `test`, `lint`, `ci` (lint+test+build), `run`, `list`, `docker-up`, `docker-down`, `publish-examples`, `clean`

## Website

Static HTML/CSS served via GitHub Pages. No JS frameworks, no CDN dependencies, no build tools.

- Base href: `/tfwatch/`
- Pages: `index.html`, `getting-started.html`, `features.html`, `docs.html`
- Supports light/dark theme toggle
- Assets: `website/assets/` (logo.svg, flow.svg, screenshots)

## CI/CD

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push + PR | Build, test (race + coverage), lint |
| `release.yml` | Push to main | release-please + GoReleaser (linux/darwin, amd64/arm64) |
| `deploy-website.yml` | website/ changes | GitHub Pages deployment |
| `commit-lint.yml` | PR | Conventional commit enforcement |

## Agent Personas

This project uses agent personas to enforce consistent, scoped AI assistance. See `prompts/agents/` for all persona definitions and `AGENTS.md` for usage instructions.

## File Ownership Boundaries

| Agent | Owns | Does NOT Touch |
|-------|------|----------------|
| PM | Scope decisions, roadmap, issue triage | Any file |
| Dev | `*.go`, `*_test.go`, `testdata/`, `go.mod`, `go.sum`, `examples/` (new features, design) | CI/CD, website, docs, Makefile, CHANGELOG.md |
| DevOps | `.github/workflows/`, `.goreleaser.yml`, `.golangci.yml`, `Makefile`, `deploy/` | Go source, docs content, website HTML/CSS, CHANGELOG.md |
| Docs | `README.md`, `CONTRIBUTING.md`, `DESIGN.md`, `SECURITY.md`, `docs/`, website copy, `doc.go` content | Go logic, CI/CD, website CSS/layout, CHANGELOG.md |
| UI | `website/*.html` (layout), `website/style.css`, `website/favicon.svg`, `website/assets/` | Go source, CI/CD, markdown docs, CHANGELOG.md |
| Fix | `*.go`, `*_test.go`, `testdata/`, `go.mod`, `go.sum`, `examples/` (bugs, review, quality) | CI/CD, website, docs, Makefile, CHANGELOG.md |

CHANGELOG.md is managed exclusively by release-please. No agent edits it.
