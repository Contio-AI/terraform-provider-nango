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

  # Optional: Set the base URL for self-hosted Nango instances.
  # Defaults to https://api.nango.dev. Can also be set via the NANGO_HOST environment variable.
  # host = "https://nango.example.com"
}
