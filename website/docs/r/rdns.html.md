---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_reverse_dns"
sidebar_current: "docs-hcloud-resource-rdns"
description: |-
  Provides a Hetzner Cloud Reverse DNS Entry to represent a reverse dns entry for a Hetzner Cloud Floating IP or a Hetzner Cloud server.
---

# hcloud_rdns

Provides a Hetzner Cloud Reverse DNS Entry to represent a reverse dns entry for a Hetzner Cloud Floating IP or a Hetzner Cloud server.

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

- `ip_address` - (Required) The ip address that has the reverse dns entry
- `server_id` - (Required) Server where the ip address belongs to (Use only one from server_id or floating_ip_id)
- `floating_ip_id` - (Required) Floating IP where the ip address belongs to (Use only one from server_id or floating_ip_id)
- `dns_ptr` - (Optional) The DNS ptr that should be set, empty if the existing reverse dns should be deleted.

## Attributes Reference

- `id` - Unique ID of the Reverse DNS Entry.
- `ip_address` - IP address that has the reverse dns entry.
- `server_id` - Server that is associated with the ip address.
- `floating_ip_id` - Floating IP that is associated with the ip address.
- `dns_ptr` - DNS PTR for the ip address.
