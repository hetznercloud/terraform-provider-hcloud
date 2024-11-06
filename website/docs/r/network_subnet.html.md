---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_network_subnet"
sidebar_current: "docs-hcloud-resource-network-subnet"
description: |-
  Provides a Hetzner Cloud Network Subnet to represent a Subnet in the Hetzner Cloud.
---

# hcloud_network_subnet

Provides a Hetzner Cloud Network Subnet to represent a Subnet in the Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_network" "mynet" {
  name     = "my-net"
  ip_range = "10.0.0.0/8"
}
resource "hcloud_network_subnet" "foonet" {
  network_id   = hcloud_network.mynet.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

```

## Argument Reference

- `network_id` - (Required, int) ID of the Network the subnet should be added to.
- `type` - (Required, string) Type of subnet. `server`, `cloud` or `vswitch`
- `ip_range` - (Required, string) Range to allocate IPs from. Must be a subnet of the ip_range of the Network and must not overlap with any other subnets or with any destinations in routes.
- `network_zone` - (Required, string) Name of network zone.
- `vswitch_id` - (Optional, int) ID of the vswitch, Required if type is `vswitch`

## Attributes Reference

- `id` - (string) ID of the Network subnet.
- `network_id` - (int) ID of the Network.
- `type` - (string) Type of subnet.
- `ip_range` - (string) Range to allocate IPs from.
- `network_zone` - (string) Name of network zone.
- `vswitch_id` - (int) ID of the vswitch, when type is `vswitch`.

## Import

Network Subnet entries can be imported using a compound ID with the following format:
`<network-id>-<ip_range>`

```
terraform import hcloud_network_subnet.mysubnet 123-10.0.0.0/24
```
