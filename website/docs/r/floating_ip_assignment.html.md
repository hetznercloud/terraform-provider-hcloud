---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_floating_ip_assignment"
sidebar_current: "docs-hcloud-resource-floating-ip-assignment"
description: |-
  Provides a Hetzner Cloud Floating IP Assignment to assign a Floating IP to a Hetzner Cloud Server.
---

# hcloud_floating_ip_assignment

Provides a Hetzner Cloud Floating IP Assignment to assign a Floating IP to a Hetzner Cloud Server. Deleting a Floating IP Assignment will unassign the Floating IP from the Server.

## Example Usage

```hcl
resource "hcloud_floating_ip_assignment" "main" {
  floating_ip_id = hcloud_floating_ip.master.id
  server_id      = hcloud_server.node1.id
}

resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx11"
  datacenter  = "fsn1-dc8"
}

resource "hcloud_floating_ip" "master" {
  type          = "ipv4"
  home_location = "nbg1"
}
```

## Argument Reference

- `floating_ip_id` - (Required, int) ID of the Floating IP.
- `server_id` - (Required, int) Server to assign the Floating IP to.

## Attributes Reference

- `id` - (int) Unique ID of the Floating IP Assignment.
- `floating_ip_id` - (int) ID of the Floating IP.
- `server_id` - (int) Server the Floating IP was assigned to.

## Import

Floating IP Assignments can be imported using the `floating_ip_id`:

```
terraform import hcloud_floating_ip_assignment.myfloatingipassignment <floating_ip_id>
```
