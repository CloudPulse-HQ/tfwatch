# Deployment Guide

## Local Observability Stack

tfwatch ships with a Docker Compose stack that includes an OTEL Collector, Prometheus, and Grafana with a pre-built dashboard.

### Start the stack

```bash
docker compose -f deploy/docker-compose.yml up -d
```

This provisions:
- **OTEL Collector** on `localhost:4317` — receives metrics from tfwatch
- **Prometheus** on `localhost:9090` — stores and queries metrics
- **Grafana** on `localhost:3000` — pre-loaded dashboard for browsing dependencies

### Run tfwatch

```bash
# Scan current directory
tfwatch

# Scan a specific repo
tfwatch --dir ./infra/prod

# Tag metrics with the pipeline phase
tfwatch --dir ./infra/prod --phase apply
```

### Stop the stack

```bash
docker compose -f deploy/docker-compose.yml down
```

## Cloud Provider Endpoints

tfwatch works with any OTEL-compatible backend. Set the `--otel-endpoint` flag (and `--otel-insecure=false` for TLS) to send metrics to your provider.

| Provider | Endpoint |
|----------|----------|
| **Local** (included) | `localhost:4317` |
| **Datadog** | `api.datadoghq.com:4317` |
| **Grafana Cloud** | `otlp-gateway-<zone>.grafana.net:4317` |
| **New Relic** | `otlp.nr-data.net:4317` |

### Example: Send to Datadog

```bash
tfwatch --dir ./infra --otel-endpoint api.datadoghq.com:4317 --otel-insecure=false
```

### Example: Send to Grafana Cloud

```bash
tfwatch --dir ./infra --otel-endpoint otlp-gateway-prod-us-east-0.grafana.net:4317 --otel-insecure=false
```

> Check your provider's docs for any required headers or authentication. Some providers require an API key passed via OTEL Collector configuration rather than directly from the client.

## CI/CD Integration

tfwatch is a single binary with no runtime dependencies, making it easy to run in CI pipelines.

### GitHub Actions

```yaml
name: tfwatch
on:
  push:
    branches: [main]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install tfwatch
        run: go install github.com/CloudPulse-HQ/tfwatch@latest

      - name: Publish metrics
        run: tfwatch --dir ./infra --phase apply --otel-endpoint ${{ secrets.OTEL_ENDPOINT }} --otel-insecure=false
```

### Tips

- Use `--phase plan` for PR builds and `--phase apply` for merge-to-main builds to distinguish environments in your dashboard.
- Run tfwatch after `terraform init` so that `.terraform.lock.hcl` is present with resolved versions.
- For multiple Terraform root modules in one repo, run tfwatch once per directory.

## Environment Variables / Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `.` | Path to Terraform configuration directory |
| `--phase` | `plan` | Terraform phase: `plan` or `apply` |
| `--otel-endpoint` | `localhost:4317` | OTEL collector gRPC endpoint |
| `--otel-insecure` | `true` | Use insecure gRPC connection (disable TLS) |
| `--list` | `false` | Print dependencies to stdout without publishing metrics |
| `--version` | | Print tfwatch version and exit |
