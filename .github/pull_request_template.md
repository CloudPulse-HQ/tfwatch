## What does this PR do?

<!-- Brief description of the change -->

## Why?

<!-- Motivation, context, or link to an issue (e.g. Fixes #123) -->

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Refactor / code quality
- [ ] Documentation
- [ ] CI/CD / tooling

## How to test

<!-- Steps to verify this change works -->

1. `make ci` (runs lint, test, build)
2. <!-- Add manual verification steps if applicable -->

## Checklist

- [ ] `make lint` passes (no new warnings)
- [ ] `make test` passes (coverage >= 80%)
- [ ] `make build` compiles
- [ ] New/changed parsing logic has table-driven tests
- [ ] Metrics verified in local Grafana (`make docker-up && make publish-examples`)

## Screenshots / metrics output

<!-- If applicable, paste tfwatch CLI output or Grafana screenshots -->
