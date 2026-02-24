terraform {
  backend "s3" {
    bucket = "mybucket"
    key    = "terraform.tfstate"
  }
}
