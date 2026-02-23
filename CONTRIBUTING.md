# Contributing to tfwatch

Thanks for your interest in contributing!

## Getting started

```bash
git clone https://github.com/CloudPulse-HQ/tfwatch.git
cd tfwatch
make build
make test
```

## Development workflow

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run `make test` to verify all tests pass
4. Run `make build` to verify the binary compiles
5. Submit a pull request

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
├── main.go           # CLI entrypoint and flag parsing
├── parser.go         # Terraform config parser (backend, modules, providers)
├── parser_test.go    # Tests
├── collector.go      # OTEL metric collection and publishing
├── deploy/           # Docker Compose observability stack
├── examples/         # Sample Terraform repos for testing
└── DESIGN.md         # Architecture and design decisions
```

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep changes focused — one feature or fix per PR
- Add tests for new functionality

## Reporting issues

Open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- tfwatch version (`tfwatch --version`)
