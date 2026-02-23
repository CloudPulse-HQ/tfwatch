provider "aws" {
  region = var.aws_region
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.2.1"
}

module "acm" {
  source  = "terraform-aws-modules/acm/aws"
  version = "5.0.1"
}
