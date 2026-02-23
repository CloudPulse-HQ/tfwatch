terraform {
  required_version = ">= 1.5.0"

  backend "s3" {
    bucket         = "acme-terraform-state"
    key            = "database/aurora/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.50"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}
