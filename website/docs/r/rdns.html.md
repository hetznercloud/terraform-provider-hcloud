---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_rdns"
sidebar_current: "docs-hcloud-resource-rdns"
description: |-
  Provides a Hetzner Cloud Reverse DNS Entry to create, modify and reset reverse dns entries for Hetzner Cloud Floating IPs or servers.
---

# hcloud_rdns

Provides a Hetzner Cloud Reverse DNS Entry to create, modify and reset reverse dns entries for Hetzner Cloud Floating IPs or servers.

## Example Usage

```hcl
resource "hcloud_server" "node1" {
  name = "node1"
  image = "debian-9"
  server_type = "cx11"
}

resource "hcloud_rdns" "master" {
  server_id = "${hcloud_server.node1.id}"
  ip_address = "${hcloud_server.node1.ipv4_address}"
  dns_ptr = "example.com"
}
```

## Argument Reference

- `dns_ptr` - (Required) The DNS address the `ip_address` should resolve to.
- `ip_address` - (Required) The IP address that should point to `dns_ptr`.
- `server_id` - (Required) The server the `ip_address` belongs to.
- `floating_ip_id` - (Required) The Floating IP the `ip_address` belongs to.

## Attributes Reference

- `id` - Unique ID of the Reverse DNS Entry.
- `dns_ptr` - DNS pointer for the IP address.
- `ip_address` - IP address.
- `server_id` - The server the IP address belongs to.
- `floating_ip_id` - The Floating IP the IP address belongs to.

## Import

Reverse DNS entries can be imported using a compound ID with the following format:

`<prefix (s for server/ f for floating ip)>-<server or floating ip ID>-<IP address>`

```
# import reverse dns entry on server with id 123, ip 192.168.100.1
terraform import hcloud_rdns.myrdns s-123-192.168.100.1

# import reverse dns entry on floating ip with id 123, ip 2001:db8::1
terraform import hcloud_rdns.myrdns f-123-2001:db8::1
```
