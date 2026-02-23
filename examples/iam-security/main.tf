provider "aws" {
  region = var.aws_region
}

provider "tls" {}

module "iam" {
  source  = "terraform-aws-modules/iam/aws"
  version = "5.40.0"
}

module "kms" {
  source  = "terraform-aws-modules/kms/aws"
  version = "3.1.1"
}
