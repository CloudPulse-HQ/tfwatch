provider "aws" {
  region = var.aws_region
}

provider "null" {}

module "lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "7.7.0"
}

module "apigateway_v2" {
  source  = "terraform-aws-modules/apigateway-v2/aws"
  version = "5.0.0"
}

module "dynamodb_table" {
  source  = "terraform-aws-modules/dynamodb-table/aws"
  version = "4.0.0"
}

module "sqs" {
  source  = "terraform-aws-modules/sqs/aws"
  version = "4.1.0"
}
