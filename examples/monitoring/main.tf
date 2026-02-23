provider "aws" {
  region = var.aws_region
}

provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
}

module "sns" {
  source  = "terraform-aws-modules/sns/aws"
  version = "6.0.0"
}

module "lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "7.14.0"
}

module "cloudwatch_log_group" {
  source  = "terraform-aws-modules/cloudwatch-log-group/aws"
  version = "3.2.0"
}
