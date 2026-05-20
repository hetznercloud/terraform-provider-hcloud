---
page_title: "hcloud_server_power_state Resource - hcloud"
subcategory: ""
description: |-
  Manages the power state of a Hetzner Cloud Server.
---

# hcloud_server_power_state

Manages the power state of a Hetzner Cloud Server.

This resource only manages whether the server is running or off. Destroying this resource removes power state management from Terraform/OpenTofu state and does not power off, power on, or delete the server.

## Example Usage

```terraform
resource "hcloud_server" "example" {
  name        = "example"
  server_type = "cx22"
  image       = "debian-12"
}

resource "hcloud_server_power_state" "example" {
  server_id = hcloud_server.example.id
  state     = "off"
}
```

## Argument Reference

The following arguments are supported:

- `server_id` - (Required, int) ID of the server whose power state should be managed.
- `state` - (Required, string) Desired power state of the server. Valid values are `running` and `off`.

## Attributes Reference

The following attributes are exported:

- `id` - (int) ID of the server whose power state is managed.
- `status` - (string) Raw status returned by the Hetzner Cloud API.

## Import

Server power state management can be imported using the server ID:

```shell
terraform import hcloud_server_power_state.example <server_id>
```
