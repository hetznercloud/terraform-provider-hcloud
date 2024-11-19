---
page_title: "Hetzner Cloud: hcloud_network"
description: |-
  Provides a Hetzner Cloud Network to represent a Network in the Hetzner Cloud.
---

# hcloud_network

Provides a Hetzner Cloud Network to represent a Network in the Hetzner Cloud.

## Example Usage

```terraform
resource "hcloud_network" "privNet" {
  name     = "my-net"
  ip_range = "10.0.1.0/24"
}
```

## Argument Reference

- `name` - (Required, string) Name of the Network to create (must be unique per project).
- `ip_range` - (Required, string) IP Range of the whole Network which must span all included subnets and route destinations. Must be one of the private ipv4 ranges of RFC1918.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.
- `delete_protection` - (Optional, bool) Enable or disable delete protection. See ["Delete Protection"](../index.html.markdown#delete-protection) in the Provider Docs for details.
- `expose_routes_to_vswitch` - (Optional, bool) Enable or disable exposing the routes to the vSwitch connection. The exposing only takes effect if a vSwitch connection is active.

## Attributes Reference

- `id` - (int) Unique ID of the network.
- `name` - (string) Name of the network.
- `ip_range` - (string) IPv4 Prefix of the whole Network.
- `labels` - (map) User-defined labels (key-value pairs)
- `delete_protection` - (bool) Whether delete protection is enabled.
- `expose_routes_to_vswitch` - (bool) Indicates if the routes from this network should be exposed to the vSwitch connection. The exposing only takes effect if a vSwitch connection is active.

## Import

Networks can be imported using its `id`:

```shell
terraform import hcloud_network.example "$NETWORK_ID"
```
