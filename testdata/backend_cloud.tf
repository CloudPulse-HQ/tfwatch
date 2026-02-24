terraform {
  cloud {
    organization = "myorg"
    workspaces {
      name = "prod"
    }
  }
}
