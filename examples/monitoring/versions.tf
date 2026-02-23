terraform {
  required_version = ">= 1.5.0"

  cloud {
    organization = "acme-corp"
    workspaces {
      name = "monitoring-prod"
    }
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70"
    }
    datadog = {
      source  = "DataDog/datadog"
      version = "~> 3.46"
    }
  }
}
