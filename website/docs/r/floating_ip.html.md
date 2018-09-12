---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_floating_ip"
sidebar_current: "docs-hcloud-resource-floating-ip-x"
description: |-
  Provides a Hetzner Cloud Floating IP to represent a publicly-accessible static IP address that can be mapped to one of your servers.
---

# hcloud_floating_ip

Provides a Hetzner Cloud Floating IP to represent a publicly-accessible static IP address that can be mapped to one of your servers.

## Example Usage

```hcl
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
- `ip_network` - IPv6 subnet. (Only set if `type` is `ipv6`)
- `labels` - User-defined labels (key-value pairs)

## Import

Floating IPs can be imported using its `id`:

```
terraform import hcloud_floating_ip.myip <id>
```
