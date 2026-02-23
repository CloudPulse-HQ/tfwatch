# Metrics Reference

## Metric Format

tfwatch emits a single gauge metric: **`terraform_dependency_version`**

The metric value is always `1`. All version and context information lives in labels, which makes it easy to query and filter in any OTEL-compatible backend.

## Labels

| Label | Description | Example |
|-------|-------------|---------|
| `backend_type` | Backend kind | `workspace`, `s3` |
| `backend_org` | Organization or bucket | `acme-corp`, `my-tf-state` |
| `backend_workspace` | Workspace or key (normalized) | `production`, `prod_vpc_tfstate` |
| `phase` | Pipeline phase | `plan`, `apply` |
| `type` | Dependency kind | `module`, `provider` |
| `dependency_name` | Name | `vpc`, `aws` |
| `dependency_source` | Registry source | `terraform-aws-modules/vpc/aws` |
| `dependency_version` | Semver version | `5.1.2` |
| `terraform_version` | Terraform CLI version | `1.9.8` |

## Use Cases

### Find repos using a vulnerable module version

```promql
terraform_dependency_version{type="module", dependency_name="vpc", dependency_version="5.0.0"}
```

### Find repos on a deprecated provider version

```promql
terraform_dependency_version{type="provider", dependency_name="aws", dependency_version!~"6.*"}
```

### Which workspaces use a specific module?

```promql
terraform_dependency_version{type="module", dependency_name="eks"}
```

### Audit all dependencies in production

```promql
terraform_dependency_version{phase="apply", backend_workspace=~".*prod.*"}
```

### Count repos by backend type

```promql
count by (backend_type) (
  count by (backend_type, backend_org, backend_workspace) (terraform_dependency_version)
)
```

### List all unique provider versions in use

```promql
count by (dependency_name, dependency_version) (
  terraform_dependency_version{type="provider"}
)
```

> All of these queries also work as Grafana dashboard filters — use the built-in filter bar to search without writing PromQL.
