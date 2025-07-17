terraform {
  required_providers {
    nango = {
      source  = "contio.ai/contio/nango"
      version = "~> 1.0"
    }
  }
}

provider "nango" {
  environment_key = "your-nango-environment-key"
}
