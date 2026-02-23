provider "aws" {
  region = var.aws_region
}

provider "random" {}

module "rds_aurora" {
  source  = "terraform-aws-modules/rds-aurora/aws"
  version = "8.5.0"
}

module "kms" {
  source  = "terraform-aws-modules/kms/aws"
  version = "3.0.0"
}
