---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_rdns"
sidebar_current: "docs-hcloud-resource-rdns"
description: |-
  Provides a Hetzner Cloud Reverse DNS Entry to create, modify and reset reverse dns entries for Hetzner Cloud Servers, Primary IPs, Floating IPs or Load Balancers.
---

# hcloud_rdns

Provides a Hetzner Cloud Reverse DNS Entry to create, modify and reset reverse dns entries for Hetzner Cloud Servers, Primary IPs, Floating IPs or Load Balancers.

## Example Usage

For servers:

```hcl
resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx11"
}

resource "hcloud_rdns" "master" {
  server_id  = hcloud_server.node1.id
  ip_address = hcloud_server.node1.ipv4_address
  dns_ptr    = "example.com"
}
```

For Primary IPs:

```hcl
resource "hcloud_primary_ip" "primary1" {
  datacenter = "nbg1-dc3"
  type       = "ipv4"
}

resource "hcloud_rdns" "primary1" {
  primary_ip_id  = hcloud_primary_ip.primary1.id
  ip_address     = hcloud_primary_ip.primary1.ip_address
  dns_ptr        = "example.com"
}
```

For Floating IPs:

```hcl
resource "hcloud_floating_ip" "floating1" {
  home_location = "nbg1"
  type          = "ipv4"
}

resource "hcloud_rdns" "floating_master" {
  floating_ip_id = "${hcloud_floating_ip.floating1.id}"
  ip_address     = "${hcloud_floating_ip.floating1.ip_address}"
  dns_ptr        = "example.com"
}
```

For Load Balancers:

```hcl
resource "hcloud_load_balancer" "load_balancer1" {
  name               = "load_balancer1"
  load_balancer_type = "lb11"
  location           = "fsn1"
}

resource "hcloud_rdns" "load_balancer_master" {
  load_balancer_id = "${hcloud_load_balancer.load_balancer1.id}"
  ip_address       = "${hcloud_load_balancer.load_balancer1.ipv4}"
  dns_ptr          = "example.com"
}
```
## Argument Reference

- `dns_ptr` - (Required, string) The DNS address the `ip_address` should resolve to.
- `ip_address` - (Required, string) The IP address that should point to `dns_ptr`.
- `server_id` - (Required, int) The server the `ip_address` belongs to. - `server_id` - (Required, int) The server the `ip_address` belongs to. Specify only one of `server_id`, `primary_ip_id`, `floating_ip_id` and `load_balancer_id`.
- `primary_ip_id` - (Required, int) The Primary IP the `ip_address` belongs to. - `server_id` - (Required, int) The server the `ip_address` belongs to. Specify only one of `server_id`, `primary_ip_id`, `floating_ip_id` and `load_balancer_id`.
- `floating_ip_id` - (Required, int) The Floating IP the `ip_address` belongs to. - `server_id` - (Required, int) The server the `ip_address` belongs to. Specify only one of `server_id`, `primary_ip_id`, `floating_ip_id` and `load_balancer_id`.
- `load_balancer_id` - (Required, int) The Load Balancer the `ip_address` belongs to. - `server_id` - (Required, int) The server the `ip_address` belongs to. Specify only one of `server_id`, `primary_ip_id`, `floating_ip_id` and `load_balancer_id`.

## Attributes Reference

- `id` - (int) Unique ID of the Reverse DNS Entry.
- `dns_ptr` - (string) DNS pointer for the IP address.
- `ip_address` - (string) IP address.
- `server_id` - (int) The server the IP address belongs to.
- `primary_ip_id` - (int) The Primary IP the IP address belongs to.
- `floating_ip_id` - (int) The Floating IP the IP address belongs to.
- `load_balancer_id` - (int) The Load Balancer the IP address belongs to.

## Import

Reverse DNS entries can be imported using a compound ID with the following format:
`<prefix (s for server/ f for floating ip / l for load balancer)>-<server, floating ip or load balancer ID>-<IP address>`

```
# import reverse dns entry on server with id 123, ip 192.168.100.1
terraform import hcloud_rdns.myrdns s-123-192.168.100.1

# import reverse dns entry on primary ip with id 123, ip 2001:db8::1
terraform import hcloud_rdns.myrdns p-123-2001:db8::1

# import reverse dns entry on floating ip with id 123, ip 2001:db8::1
terraform import hcloud_rdns.myrdns f-123-2001:db8::1

# import reverse dns entry on load balancer with id 123, ip 2001:db8::1
terraform import hcloud_rdns.myrdns l-123-2001:db8::1
```
