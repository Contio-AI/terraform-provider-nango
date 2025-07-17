terraform {
  required_providers {
    nango = {
      source  = "registry.terraform.io/contio/nango"
      version = "~> 1.0"
    }
  }
}

provider "nango" {
  environment_key = "your-nango-environment-key"
}
