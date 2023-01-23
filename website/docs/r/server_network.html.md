---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_server_network"
sidebar_current: "docs-hcloud-resource-server-network"
description: |-
  Provides a Hetzner Cloud Server Network to represent a private network on a server in the Hetzner Cloud.
---

# hcloud_server_network

 Provides a Hetzner Cloud Server Network to represent a private network on a server in the Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx11"
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

resource "hcloud_server_network" "srvnetwork" {
  server_id  = hcloud_server.node1.id
  network_id = hcloud_network.mynet.id
  ip         = "10.0.1.5"
}
```

## Argument Reference

- `server_id` - (Required, int) ID of the server.
- `alias_ips` - (Optional, list[string]) Additional IPs to be assigned
  to this server.
- `network_id` - (Optional, int) ID of the network which should be added
  to the server. Required if `subnet_id` is not set. Successful creation
  of the resource depends on the existence of a subnet in the Hetzner
  Cloud Backend. Using `network_id` will not create an explicit
  dependency between server and subnet. Therefore `depends_on` may need
  to be used. Alternatively the `subnet_id` property can be used, which
  will create an explicit dependency between `hcloud_server_network` and
  the existence of a subnet.
- `subnet_id` - (Optional, string) ID of the sub-network which should be
  added to the Server. Required if `network_id` is not set.
  *Note*: if the `ip` property is missing, the Server is currently added
  to the last created subnet.
- `ip` - (Optional, string) IP to request to be assigned to this server.
  If you do not provide this then you will be auto assigned an IP
  address.

## Attributes Reference

- `id` - (string) ID of the server network.
- `network_id` - (int) ID of the network.
- `server_id` - (int) ID of the server.
- `ip` - (string) IP assigned to this server.
- `alias_ips` - (list[string]) Additional IPs assigned to this server.

## Import

Server Network entries can be imported using a compound ID with the following format:
`<server-id>-<network-id>`

```
terraform import hcloud_server_network.myservernetwork 123-654
```
