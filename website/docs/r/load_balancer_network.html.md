---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_load_balancer_network"
sidebar_current: "docs-hcloud-resource-load_balancer-network"
description: |-
  Provides a Hetzner Cloud Load Balancer Network to represent a private network on a Load Balancer in the Hetzner Cloud.
---

# hcloud_load_balancer_network

 Provides a Hetzner Cloud Load Balancer Network to represent a private network on a Load Balancer in the Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_load_balancer" "lb1" {
  name = "lb1"
  load_balancer_type = "lb11"
  network_zone = "eu-central"
}

resource "hcloud_network" "mynet" {
  name = "my-net"
  ip_range = "10.0.0.0/8"
}

resource "hcloud_network_subnet" "foonet" {
  network_id = hcloud_network.mynet.id
  type = "cloud"
  network_zone = "eu-central"
  ip_range   = "10.0.1.0/24"
}

resource "hcloud_load_balancer_network" "srvnetwork" {
  load_balancer_id = hcloud_load_balancer.lb1.id
  network_id = hcloud_network.mynet.id
  ip = "10.0.1.5"
}
```

## Argument Reference

- `load_balancer_id` - (Required, int) ID of the Load Balancer.
- `network_id` - (Optional, int) ID of the network which should be added
  to the Load Balancer. Required if `subnet_id` is not set. Successful
  creation of the resource depends on the existence of a subnet in the
  Hetzner Cloud Backend. Using `network_id` will not create an explicit
  dependency between load balancer and subnet. It is thus better to use
  the `subnet_id` property. This property is deprecated.
- `subnet_id` - (Optional, string) ID of the sub-network which should be
  added to the Load Balancer. Required if `network_id` is not set.
- `ip` - (Optional, string) IP to request to be assigned to this Load
  Balancer. If you do not provide this then you will be auto assigned an
  IP address.
- `enable_public_interface` - (Optional, bool) Enable or disable the
  Load Balancers public interface. Default: `true`

## Attributes Reference

- `id` - (string) ID of the Load Balancer network.
- `network_id` - (int) ID of the network.
- `load_balancer_id` - (int) ID of the Load Balancer.
- `ip` - (string) IP assigned to this Load Balancer.
