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
  name        = "node1"
  image       = "debian-11"
  server_type = "cx11"
}

resource "hcloud_floating_ip" "master" {
  type      = "ipv4"
  server_id = hcloud_server.node1.id
}
```

## Argument Reference

- `type` - (Required, string) Type of the Floating IP. `ipv4` `ipv6`
- `name` - (Optional, string) Name of the Floating IP.
- `server_id` - (Optional, int) Server to assign the Floating IP to.
- `home_location` - (Optional, string) Name of home location (routing is optimized for that location). Optional if server_id argument is passed.
- `description` - (Optional, string) Description of the Floating IP.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.
- `delete_protection` - (Optional, bool) Enable or disable delete protection.

## Attributes Reference

- `id` - (int) Unique ID of the Floating IP.
- `type` - (string) Type of the Floating IP.
- `name` - (string) Name of the Floating IP.
- `server_id` - (int) Server to assign the Floating IP is assigned to.
- `home_location` - (string) Home location.
- `description` - (string) Description of the Floating IP.
- `ip_address` - (string) IP Address of the Floating IP.
- `ip_network` - (string) IPv6 subnet. (Only set if `type` is `ipv6`)
- `labels` - (map) User-defined labels (key-value pairs)
- `delete_protection` - (bool) Whether delete protection is enabled.

## Import

Floating IPs can be imported using its `id`:

```
terraform import hcloud_floating_ip.myip id
```
