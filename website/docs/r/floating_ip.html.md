---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_floating_ip"
sidebar_current: "docs-hcloud-resource-floating-ip"
description: |-
  Provides a Hetzner Cloud Floating IP to represent a publicly-accessible static IP addresses that can be mapped to one of your Servers.
---

# hcloud_floating_ip

Provides a Hetzner Cloud Floating IP to represent a publicly-accessible static IP addresses that can be mapped to one of your Servers.

## Example Usage

```
resource "hcloud_server" "node1" {
  name = "node1"
  image = "debian-9"
  server_type = "cx11"
}

resource "hcloud_floating_ip" "master" {
  type = "ipv4"
  server_id = "${hcloud_server.node1.id}"
}
```

## Argument Reference

- `type` - (Required) Type of the Floating IP. `ipv4` `ipv6`
- `server_id` - (Optional) Server to assign the Floating IP to.
- `home_location` - (Optional) Home location (routing is optimized for that location). Optional if server_id argument is passed.
- `description` - (Optional) Description of the Floating IP.

## Attributes Reference

- `id` - Unique ID of the Floating IP.
- `type` - Type of the Floating IP.
- `server_id` - Server to assign the Floating IP is assigned to.
- `home_location` - Home location.
- `description` - Description of the Floating IP.
- `ip_address` - IP Address of the Floating IP.

## Import

Floating IPs can be imported using its `id`:

```
terraform import hcloud_floating_ip.myip <id>
```
