# Example Terraform Repos

This directory contains sample Terraform configurations for testing tfwatch. Each subdirectory simulates a real Terraform repo with a backend configuration, modules, and providers.

## Repos

| Directory | Backend | Modules | Description |
|-----------|---------|---------|-------------|
| `app-frontend` | Terraform Cloud | Various | Frontend application infra |
| `data-platform` | S3 | Various | Data/analytics platform |
| `database` | S3 | RDS, ElastiCache | Database infrastructure |
| `dns-cdn` | S3 | CloudFront, Route53 | DNS and CDN setup |
| `eks-cluster` | S3 | EKS, VPC | Kubernetes cluster |
| `iam-security` | Terraform Cloud | IAM modules | IAM roles and policies |
| `monitoring` | Terraform Cloud | CloudWatch, SNS | Monitoring and alerting |
| `multi-cloud` | Terraform Cloud | AWS + GCP modules | Multi-cloud setup |
| `networking` | Terraform Cloud | VPC, subnets | Core networking |
| `serverless` | S3 | Lambda, API Gateway | Serverless workloads |

## Usage

List dependencies from a single example:

```bash
tfwatch --list --dir ./examples/eks-cluster
```

Publish all examples to the local OTEL stack:

```bash
make publish-examples
```
