---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_network"
sidebar_current: "docs-hcloud-resource-network"
description: |-
  Provides a Hetzner Cloud Network to represent a private network in the Hetzner Cloud.
---

# hcloud_network

 Provides a Hetzner Cloud Network to represent a private network in the Hetzner Cloud.

## Example Usage

For servers:

```hcl
resource "hcloud_server" "node1" {
  name = "node1"
  image = "debian-9"
  server_type = "cx11"
}

resource "hcloud_network" "privNet" {
  name = "my-net"
  ip_range = "10.0.1.0/24"
  dns_ptr = "example.com"
}
```

For Floating IPs:

```hcl
resource "hcloud_floating_ip" "floating1" {
  home_location = "nbg1"
  type = "ipv4"
}

resource "hcloud_rdns" "floating_master" {
  floating_ip_id = "${hcloud_floating_ip.floating1.id}"
  ip_address = "${hcloud_floating_ip.floating1.ip_address}"
  dns_ptr = "example.com"
}
```
## Argument Reference

- `name` - (Required, string) Name of the network to create (must be unique per project).
- `ip_range` - (Required, string) IP Range of the whole Network which must span all included subnets and route destionations. Must be one of the private ipv4 ranges of RFC1918.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.

## Attributes Reference

- `id` - (int) Unique ID of the network.
- `name` - (string) Name of the network.
- `ip_range` - (string) IPv4 Prefix of the whole Network.
- `labels` - (map) User-defined labels (key-value pairs)

## Import

Networks can be imported using its `id`:

```
terraform import hcloud_network.myip <id>
```

