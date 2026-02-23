provider "aws" {
  region = var.aws_region
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.1.2"
}

module "transit_gateway" {
  source  = "terraform-aws-modules/transit-gateway/aws"
  version = "2.12.2"
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "5.1.0"
}
