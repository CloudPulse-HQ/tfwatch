terraform {
  required_version = ">= 1.5.0"

  cloud {
    organization = "acme-corp"
    workspaces {
      name = "iam-security-prod"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.65"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}
