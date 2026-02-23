terraform {
  required_version = ">= 1.5.0"

  cloud {
    organization = "acme-corp"
    workspaces {
      name = "app-frontend-prod"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.82"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.40"
    }
  }
}
