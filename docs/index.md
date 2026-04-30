---
page_title: "Clouding Provider"
description: |-
  Use the Clouding provider to manage Clouding.io infrastructure with Terraform.
---

# Clouding Provider

The **Clouding** provider lets you manage [Clouding.io](https://clouding.io) resources
using Terraform. It communicates with the Clouding.io REST API using your API key.

This is a **community provider** maintained by [ideaaiplus](https://github.com/ideaaiplus),
not an official Clouding.io product.

## Example Usage

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

variable "clouding_api_key" {
  type      = string
  sensitive = true
}
```

> **Security:** never hardcode your API key. Use a variable, environment variable,
> or a secrets manager (Vault, AWS Secrets Manager, etc.).

## Authentication

Obtain your API key from the [Clouding.io panel](https://panel.clouding.io):
**Account → API → Generate key**.

The provider uses the `X-API-KEY` header on every request.

## Provider Configuration

### Required

| Argument | Type | Description |
|---|---|---|
| `api_key` | string | Clouding.io API key. Marked sensitive — will not appear in plan output. |

## Resources

| Resource | Description |
|---|---|
| [clouding_server](resources/server.md) | Manages a Clouding.io virtual server. |

## Environment Variables

As an alternative to the HCL `api_key` argument, you can pass the key via
a Terraform variable and the `TF_VAR_` prefix convention:

```bash
export TF_VAR_clouding_api_key="your-api-key"
terraform plan
```
