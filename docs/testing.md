# Testing Guide

## Example Repos

tfwatch includes 10 sample Terraform configurations under `examples/` that simulate real-world repos with different backends, modules, and providers.

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

## Quick Test: List Dependencies

Print dependencies from a single example repo without publishing metrics:

```bash
make list
# or
tfwatch --list --dir ./examples/eks-cluster
```

## Generate Sample Data

### 1. Start the local stack

```bash
make docker-up
```

### 2. Publish all example repos

```bash
make publish-examples
```

This loops through every directory under `examples/` and publishes metrics for each one.

### 3. View in Grafana

Open [http://localhost:3000](http://localhost:3000). The dashboard is auto-provisioned and will show data from all 10 example repos, including:

- Total repos, dependencies, unique providers, and unique modules
- Backend type distribution (Terraform Cloud vs S3)
- Searchable dependency tables
- Provider and module version comparison across repos

### 4. Stop the stack

```bash
make docker-down
```

## Running Unit Tests

```bash
make test
# or
go test -v ./...
```

## Adding Your Own Test Repos

To add a new example repo:

1. Create a new directory under `examples/`:
   ```bash
   mkdir examples/my-new-repo
   ```

2. Add Terraform files with a backend configuration:
   ```hcl
   # examples/my-new-repo/main.tf
   terraform {
     backend "s3" {
       bucket = "my-state-bucket"
       key    = "my-new-repo/terraform.tfstate"
       region = "us-east-1"
     }
   }

   module "vpc" {
     source  = "terraform-aws-modules/vpc/aws"
     version = "5.1.2"
   }
   ```

3. Run `terraform init` in the directory to generate `.terraform.lock.hcl` (or create `modules.json` manually under `.terraform/modules/`).

4. Test it:
   ```bash
   tfwatch --list --dir ./examples/my-new-repo
   ```

## Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make test` | Run unit tests |
| `make list` | List deps from `examples/eks-cluster` |
| `make docker-up` | Start local observability stack |
| `make publish-examples` | Publish all example repos to local stack |
| `make docker-down` | Stop the stack |
| `make clean` | Remove build artifacts |
