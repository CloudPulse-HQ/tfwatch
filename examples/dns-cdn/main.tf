provider "aws" {
  region = var.aws_region
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

module "route53" {
  source  = "terraform-aws-modules/route53/aws"
  version = "4.0.0"
}

module "acm" {
  source  = "terraform-aws-modules/acm/aws"
  version = "4.5.0"
}

module "cloudfront" {
  source  = "terraform-aws-modules/cloudfront/aws"
  version = "3.3.0"
}
