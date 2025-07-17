# Terraform Provider for Nango

This Terraform provider allows you to manage [Nango](https://www.nango.dev/) integrations as infrastructure as code.

## Features

- **Integration Management**: Create, read, update, and delete Nango integrations
- **OAuth Configuration**: Configure OAuth2 credentials and scopes for various providers
- **Data Sources**: Query existing integrations in your Nango environment

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23 (for development)
- A Nango account with an environment key

## Installation

### Terraform Registry (Coming Soon)

```hcl
terraform {
  required_providers {
    nango = {
      source  = "contio.ai/contio/nango"
      version = "~> 1.0"
    }
  }
}
```

### Local Development

1. Clone this repository
2. Build the provider: `make build`
3. Install locally: `make install`

## Usage

### Provider Configuration

```hcl
provider "nango" {
  environment_key = "your-nango-environment-key"
}
```

### Creating an Integration

```hcl
resource "nango_integration" "google" {
  unique_key     = "platform-google"
  display_name   = "Google"
  nango_provider = "google"

  credentials = {
    client_id     = var.google_client_id
    client_secret = var.google_client_secret
    type          = "OAUTH2"
    scopes = [
      "https://www.googleapis.com/auth/userinfo.profile",
      "https://www.googleapis.com/auth/userinfo.email",
      "https://www.googleapis.com/auth/calendar.readonly"
    ]
  }
}
```

### Querying Integrations

```hcl
data "nango_integrations" "all" {}

output "integrations" {
  value = data.nango_integrations.all.integrations
}
```

## Resources

### `nango_integration`

Manages a Nango integration configuration.

#### Arguments

- `unique_key` (Required) - Unique identifier for the integration
- `display_name` (Required) - Human-readable name for the integration
- `nango_provider` (Required) - The Nango provider type (e.g., "google", "microsoft")
- `credentials` (Required) - OAuth credentials configuration
  - `client_id` (Required) - OAuth client ID
  - `client_secret` (Required) - OAuth client secret
  - `type` (Required) - Credential type (typically "OAUTH2")
  - `scopes` (Required) - List of OAuth scopes

#### Attributes

- `updated_at` - Timestamp of last update

## Data Sources

### `nango_integrations`

Retrieves all integrations in your Nango environment.

#### Attributes

- `integrations` - List of integration objects with the same structure as the resource

## Examples

See the [examples](./examples/) directory for complete configuration examples including:

- Google OAuth integration
- Microsoft OAuth integration
- Multiple service-specific integrations

## Development

### Building the Provider

```bash
make build
```

### Running Tests

```bash
make test
make testacc
```

### Generating Documentation

```bash
make generate
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test` and `make testacc`
6. Submit a pull request

## License

This project is licensed under the MPL-2.0 License - see the LICENSE file for details.

## Support

For issues related to this Terraform provider, please open an issue on GitHub.

For Nango-specific questions, refer to the [Nango documentation](https://docs.nango.dev/).
