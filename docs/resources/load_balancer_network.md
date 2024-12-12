---
page_title: "Hetzner Cloud: hcloud_load_balancer_network"
description: |-
  Provides a Hetzner Cloud Load Balancer Network to represent a private network on a Load Balancer in the Hetzner Cloud.
---

# hcloud_load_balancer_network

Provides a Hetzner Cloud Load Balancer Network to represent a private network on a Load Balancer in the Hetzner Cloud.

## Example Usage

```terraform
resource "hcloud_load_balancer" "lb1" {
  name               = "lb1"
  load_balancer_type = "lb11"
  network_zone       = "eu-central"
}

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

resource "hcloud_load_balancer_network" "srvnetwork" {
  load_balancer_id = hcloud_load_balancer.lb1.id
  network_id       = hcloud_network.mynet.id
  ip               = "10.0.1.5"

  # **Note**: the depends_on is important when directly attaching the
  # server to a network. Otherwise Terraform will attempt to create
  # server and sub-network in parallel. This may result in the server
  # creation failing randomly.
  depends_on = [
    hcloud_network_subnet.foonet
  ]
}
```

## Argument Reference

- `load_balancer_id` - (Required, int) ID of the Load Balancer.
- `network_id` - (Optional, int) ID of the network which should be added
  to the Load Balancer. Required if `subnet_id` is not set. Successful
  creation of the resource depends on the existence of a subnet in the
  Hetzner Cloud Backend. Using `network_id` will not create an explicit
  dependency between the Load Balancer and the subnet. Therefore
  `depends_on` may need to be used. Alternatively the `subnet_id`
  property can be used, which will create an explicit dependency between
  `hcloud_load_balancer_network` and the existence of a subnet.
- `subnet_id` - (Optional, string) ID of the sub-network which should be
  added to the Load Balancer. Required if `network_id` is not set.
  _Note_: if the `ip` property is missing, the Load Balancer is
  currently added to the last created subnet.
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

## Import

Load Balancer Network entries can be imported using a compound ID with the following format:
`<load-balancer-id>-<network-id>`

```shell
terraform import hcloud_load_balancer_network.example "$LOAD_BALANCER_ID-$NETWORK_ID"
```
