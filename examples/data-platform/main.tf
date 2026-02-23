provider "aws" {
  region = var.aws_region
}

provider "random" {}

module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.0"
}

module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "6.8.0"
}

module "lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "7.10.0"
}
