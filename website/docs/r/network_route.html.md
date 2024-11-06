---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_network_route"
sidebar_current: "docs-hcloud-resource-network-route"
description: |-
  Provides a Hetzner Cloud Network Route to represent a Network route in the Hetzner Cloud.
---

# hcloud_network_route

Provides a Hetzner Cloud Network Route to represent a Network route in the Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_network" "mynet" {
  name     = "my-net"
  ip_range = "10.0.0.0/8"
}
resource "hcloud_network_route" "privNet" {
  network_id  = hcloud_network.mynet.id
  destination = "10.100.1.0/24"
  gateway     = "10.0.1.1"
}

```

## Argument Reference

- `network_id` - (Required, int) ID of the Network the route should be added to.
- `destination` - (Required, string) Destination network or host of this route. Must be a subnet of the ip_range of the Network. Must not overlap with an existing ip_range in any subnets or with any destinations in other routes or with the first ip of the networks ip_range or with 172.31.1.1.
- `gateway` - (Required, string) Gateway for the route. Cannot be the first ip of the networks ip_range and also cannot be 172.31.1.1 as this IP is being used as a gateway for the public network interface of servers.

## Attributes Reference

- `id` - (int) Unique ID of the Network route.
- `network_id` - (int) ID of the Network.
- `destination` - (string) Destination of this route.
- `gateway` - (string) Gateway of the route.

## Import

Network Route entries can be imported using a compound ID with the following format:
`<network-id>-<destination>`

```
terraform import hcloud_network_route.myroute 123-10.0.0.0/16
```
