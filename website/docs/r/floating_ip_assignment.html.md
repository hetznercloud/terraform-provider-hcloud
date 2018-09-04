---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_floating_ip_association"
sidebar_current: "docs-hcloud-resource-floating-ip-association"
description: |-
  Provides a Hetzner Cloud Floating IP Association to associate a Floating IP to a Hetzner Cloud Server.
---

# hcloud_floating_ip_assignment

Provides a Hetzner Cloud Floating IP Association to associate a Floating IP to a Hetzner Cloud Server. Deleting a Floating IP Association disassociates the Floating IP from the Server.

## Example Usage

```hcl
resource "hcloud_floating_ip_assignment" "main" {
  floating_ip_id = "${hcloud_floating_ip.master.id}"
  server_id = "${hcloud_server.node1.id}"
}

resource "hcloud_server" "node1" {
  name = "node1"
  image = "debian-9"
  server_type = "cx11"
  datacenter = "fsn1-dc8"
}

resource "hcloud_floating_ip" "master" {
  type = "ipv4"
  home_location = "nbg1"
}
```

## Argument Reference

- `floating_ip_id` - (Required) ID of the Floating IP.
- `server_id` - (Required) Server to assign the Floating IP to.

## Attributes Reference

- `id` - Unique ID of the Floating IP Assignment.
- `floating_ip_id` - ID of the Floating IP.
- `server_id` - Server the Floating IP was assigned to.
