terraform {
  required_version = ">= 1.5.0"

  backend "s3" {
    bucket         = "acme-terraform-state"
    key            = "serverless/api/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-locks"
    encrypt        = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.82"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
  }
}
