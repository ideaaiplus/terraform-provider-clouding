# Terraform Provider for Clouding.io

[![Tests](https://github.com/ideaaiplus/terraform-provider-clouding/actions/workflows/test.yml/badge.svg)](https://github.com/ideaaiplus/terraform-provider-clouding/actions/workflows/test.yml)
[![Registry](https://img.shields.io/badge/Terraform%20Registry-ideaaiplus%2Fclouding-blue)](https://registry.terraform.io/providers/ideaaiplus/clouding/latest)

A community Terraform provider for managing [Clouding.io](https://clouding.io) infrastructure.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (to build from source)

## Usage

```hcl
terraform {
  required_providers {
    clouding = {
      source  = "ideaaiplus/clouding"
      version = "~> 0.1"
    }
  }
}

provider "clouding" {
  api_key = var.clouding_api_key
}
```

> **Tip:** set `CLOUDING_API_KEY` as a Terraform variable or use a `.tfvars` file — never hardcode secrets.

## Provider Configuration

| Argument | Required | Description |
|---|---|---|
| `api_key` | yes | Clouding.io API key (`X-API-KEY`). Obtain it from your [Clouding panel](https://panel.clouding.io). |

## Resources

### `clouding_server`

Manages a Clouding.io server.

```hcl
resource "clouding_server" "web" {
  name      = "web-01"
  hostname  = "web-01.example.com"
  image_id  = "ubuntu-22.04-lts"
  flavor_id = "2vcpu-4gb"

  # optional
  ssh_key_id = "my-key-id"
}

output "server_id" {
  value = clouding_server.web.id
}
```

#### Schema

| Attribute | Type | Required | Computed | Description |
|---|---|---|---|---|
| `id` | string | — | yes | Server ID assigned by the API |
| `name` | string | yes | — | Human-readable server name |
| `hostname` | string | yes | — | Fully-qualified hostname |
| `image_id` | string | yes | — | OS image ID (forces replacement on change) |
| `flavor_id` | string | yes | — | Hardware flavor ID (forces replacement on change) |
| `ssh_key_id` | string | no | — | SSH key ID to inject (forces replacement on change) |

#### Import

Existing servers can be imported using their Clouding.io ID:

```bash
terraform import clouding_server.web <server-id>
```

## Building from Source

```bash
git clone https://github.com/ideaaiplus/terraform-provider-clouding
cd terraform-provider-clouding
go build -o terraform-provider-clouding .
```

For local development with Terraform, add to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "ideaaiplus/clouding" = "/path/to/terraform-provider-clouding"
  }
  direct {}
}
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first.

## License

[Mozilla Public License 2.0](LICENSE)
