# Design

This document covers the architectural decisions behind tfwatch: how backends are detected, why the metric format uses a single gauge, and how the label schema is structured.

## Architecture

tfwatch is a single-binary CLI tool. It reads Terraform configuration files, resolves dependencies, and publishes them as OpenTelemetry metrics.

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  .tf files   │────▶│   Parser    │────▶│  Collector   │────▶│ OTEL gRPC   │
│  .lock.hcl   │     │  (HCL v2)  │     │  (SDK Gauge) │     │  Endpoint   │
│  modules.json│     └─────────────┘     └──────────────┘     └─────────────┘
└─────────────┘
```

### Components

**Parser** (`parser.go`) — Reads Terraform files using the HashiCorp HCL v2 library. Responsible for:
- Scanning `*.tf` files for `terraform {}` blocks to detect the backend
- Reading `.terraform/modules/modules.json` for module versions
- Reading `.terraform.lock.hcl` for provider versions
- Running `terraform init` automatically if lock files are missing

**Collector** (`collector.go`) — Creates an OpenTelemetry `Int64Gauge` and records one data point per dependency with labels describing the source repo, backend, and version.

**Main** (`main.go`) — Wires the parser and collector together. Handles flag parsing, OTEL SDK initialization (with gRPC exporter), and the `--list` mode for local debugging.

## Backend Auto-Detection

tfwatch scans all `*.tf` files in the target directory for `terraform {}` blocks. It checks for two backend types in order:

1. **Terraform Cloud / Enterprise** — looks for a `cloud {}` block:
   ```hcl
   terraform {
     cloud {
       organization = "acme-corp"
       workspaces { name = "production" }
     }
   }
   ```
   Extracts `organization` and `workspaces.name`. Sets `backend_type = "workspace"`.

2. **S3** — looks for a `backend "s3" {}` block:
   ```hcl
   terraform {
     backend "s3" {
       bucket = "my-state"
       key    = "prod/vpc/terraform.tfstate"
     }
   }
   ```
   Extracts `bucket` and `key`. The key is normalized by replacing `/` with `_` to produce cleaner label values (e.g., `prod_vpc_terraform.tfstate`). Sets `backend_type = "s3"`.

The first match wins — if a file has both `cloud {}` and `backend "s3" {}`, the cloud block takes precedence. This matches Terraform's own behavior where `cloud {}` overrides legacy backend blocks.

### Why auto-detect?

Requiring users to pass `--backend-type`, `--org`, `--workspace` flags would add friction and create opportunities for mismatch between flags and actual config. By reading the same `.tf` files that Terraform reads, tfwatch stays in sync automatically.

## Metric Format

### One gauge, many labels

tfwatch emits a single metric: `terraform_dependency_version` (Int64Gauge, value always `1`).

All version and context information is encoded in labels rather than in the metric value. This was a deliberate choice:

- **Semver is not a number.** Version strings like `5.1.2` don't work as gauge values — you can't average or sum them meaningfully. Encoding version as a label preserves the full string.
- **One metric simplifies querying.** Instead of separate metrics for modules vs providers (or per-backend), a single metric with `type`, `backend_type`, etc. lets you slice any way you want with label selectors.
- **Cardinality is bounded.** Each unique combination of (repo, dependency, version) is one time series. For a typical org with 50 repos and ~10 dependencies each, that's ~500 series — well within Prometheus limits.

### Why gauge and not counter?

A counter would require tracking "new versions" over time. tfwatch is a point-in-time scanner — it reports what's deployed right now. A gauge with value `1` means "this dependency exists at this version in this repo." When a version changes, the old series disappears and a new one appears.

## Label Schema

| Label | Source | Purpose |
|-------|--------|---------|
| `backend_type` | Parsed from `.tf` | Distinguish Cloud vs S3 repos |
| `backend_org` | `organization` or `bucket` | Group by org/account |
| `backend_workspace` | `workspaces.name` or `key` | Identify the specific deployment |
| `phase` | `--phase` flag | Separate plan-time from apply-time scans |
| `type` | Derived | `module` or `provider` |
| `dependency_name` | Module key or provider basename | Human-readable short name |
| `dependency_source` | Registry path | Full source for disambiguation |
| `dependency_version` | Lock file / modules.json | The resolved version string |
| `terraform_version` | `terraform version -json` | Track Terraform CLI drift |

### Unified org/workspace labels

Both `backend_org` and `backend_workspace` are used regardless of backend type. For Terraform Cloud, org = organization and workspace = workspace name. For S3, org = bucket and workspace = normalized key. This allows queries like `{backend_org="acme-corp"}` to work across backend types without knowing which backend a repo uses.

## Dependency Resolution

### Modules

Modules are read from `.terraform/modules/modules.json`, which Terraform generates during `init`. This file contains the resolved source and version for every module in the configuration. The root module entry (empty `Key`) is skipped.

### Providers

Providers are read from `.terraform.lock.hcl`, also generated during `init`. This file uses HCL syntax with `provider` blocks containing the registry source path and resolved version. The provider's short name is derived from the last segment of the source path (e.g., `registry.terraform.io/hashicorp/aws` → `aws`).

### Auto-init

If `.terraform.lock.hcl` is missing, tfwatch runs `terraform init` automatically. This ensures the tool works on fresh clones without requiring users to remember a setup step. In CI/CD pipelines, init typically runs before tfwatch anyway.
