terraform {
  required_version = ">= 1.5.0"

  backend "s3" {
    bucket         = "acme-terraform-state"
    key            = "multi-cloud/infra/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.55"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 6.10"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.10"
    }
  }
}
