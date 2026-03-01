# Contributing to tfwatch

Thanks for your interest in contributing!

## Getting started

```bash
git clone https://github.com/yashn3yx/tfwatch.git
cd tfwatch
make build
make test
```

## Development workflow

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run `make lint` to check for lint errors
4. Run `make test` to verify all tests pass
5. Run `make build` to verify the binary compiles
6. Submit a pull request

## Running locally

Start the observability stack:

```bash
make docker-up
```

Publish example repos:

```bash
make publish-examples
```

View metrics at [http://localhost:3000](http://localhost:3000) (Grafana) or [http://localhost:9090](http://localhost:9090) (Prometheus).

## Project structure

```
.
├── main.go             # CLI entrypoint and flag parsing
├── parser.go           # Terraform config parser (backend, modules, providers)
├── collector.go        # OTEL metric collection and publishing
├── doc.go              # Package-level godoc
├── *_test.go           # Tests (table-driven, ~82% coverage)
├── testdata/           # Static test fixtures (HCL, JSON)
├── .golangci.yml       # Linter configuration
├── deploy/             # Docker Compose observability stack
├── examples/           # Sample Terraform repos for testing
└── DESIGN.md           # Architecture and design decisions
```

## Code style

- Run `make lint` before submitting — this runs `golangci-lint` (covers `gofmt`, `go vet`, `staticcheck`, and more)
- Install golangci-lint: `brew install golangci-lint` or see [golangci-lint.run](https://golangci-lint.run/welcome/install/)
- Keep changes focused — one feature or fix per PR
- Add tests for new functionality

## AI-assisted development

This project includes structured AI agent personas for consistent assistance. See [`AGENTS.md`](AGENTS.md) for the full guide.

**Quick start:**
- Agent persona definitions live in `prompts/agents/`
- Project context is in `prompts/CONTEXT.md`
- AI safety guardrails are in `prompts/GUARDRAILS.md`
- Claude Code users get slash commands (`/dev`, `/fix`, `/docs`, etc.) automatically
- For other tools (Cursor, Copilot, Windsurf, Aider), see `AGENTS.md` for setup instructions

**Key rule:** AI-generated code must pass the same quality gates as human code — `make lint`, `make test` (80%+ coverage), `make build`.

## Reporting issues

Open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- tfwatch version (`tfwatch --version`)
