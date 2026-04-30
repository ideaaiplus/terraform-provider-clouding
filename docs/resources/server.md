---
page_title: "clouding_server Resource - Clouding Provider"
description: |-
  Manages a Clouding.io virtual server.
---

# clouding_server (Resource)

Manages a Clouding.io virtual server.

Servers are billed from the moment they are created. Destroying the resource
removes the server permanently.

## Example Usage

### Basic server

```hcl
resource "clouding_server" "web" {
  name      = "web-01"
  hostname  = "web-01.example.com"
  image_id  = "ubuntu-22.04-lts"
  flavor_id = "2vcpu-4gb"
}
```

### Server with SSH key

```hcl
resource "clouding_server" "web" {
  name       = "web-01"
  hostname   = "web-01.example.com"
  image_id   = "ubuntu-22.04-lts"
  flavor_id  = "2vcpu-4gb"
  ssh_key_id = "my-ssh-key-id"
}
```

### Multiple servers

```hcl
resource "clouding_server" "app" {
  count     = 3
  name      = "app-${count.index + 1}"
  hostname  = "app-${count.index + 1}.example.com"
  image_id  = "ubuntu-22.04-lts"
  flavor_id = "4vcpu-8gb"
}
```

## Argument Reference

### Required

| Argument | Type | Description |
|---|---|---|
| `name` | string | Human-readable name for the server. Can be updated in-place. |
| `hostname` | string | Fully-qualified hostname assigned to the server. Can be updated in-place. |
| `image_id` | string | ID of the OS image to deploy. **Forces replacement if changed.** |
| `flavor_id` | string | ID of the hardware flavor (vCPU + RAM profile). **Forces replacement if changed.** |

### Optional

| Argument | Type | Description |
|---|---|---|
| `ssh_key_id` | string | ID of the SSH public key to inject during provisioning. **Forces replacement if changed.** |

## Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | string | Unique server identifier assigned by the Clouding.io API. |

## Import

An existing server can be imported into Terraform state using its Clouding.io server ID:

```bash
terraform import clouding_server.web <server-id>
```

After importing, run `terraform plan` to verify the state matches the live configuration.

## Notes

### Immutable fields
`image_id`, `flavor_id`, and `ssh_key_id` are set only at creation time.
Changing any of them forces the creation of a new server and destruction of the old one.

### Mutable fields
`name` and `hostname` are updated in-place via `PUT /v1/servers/{id}` without
requiring resource replacement.

### Drift detection
If a server is deleted outside of Terraform (e.g. from the Clouding panel),
the next `terraform plan` will detect the drift and propose re-creating the resource.

### API field mapping
The provider handles the translation between HCL attribute names and the Clouding.io
REST API JSON structure:

| HCL attribute | API JSON path |
|---|---|
| `image_id` | `volume.id` (with `volume.source = "Image"`) |
| `flavor_id` | `flavorId` |
| `ssh_key_id` | `accessConfiguration.sshKeyId` |
