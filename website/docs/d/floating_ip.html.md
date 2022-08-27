---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_floating_ip"
sidebar_current: "docs-hcloud-datasource-floating-ip"
description: |-
  Provides details about a specific Hetzner Cloud Floating IP.
---

# Data Source: hcloud_floating_ip

Provides details about a Hetzner Cloud Floating IP.

This resource can be useful when you need to determine a Floating IP ID based on the IP address.

## Example Usage

# Data Source: hcloud_floating_ip
Provides details about a Hetzner Cloud Floating IP.
This resource can be useful when you need to determine a Floating IP ID based on the IP address.

## Example Usage
```hcl
data "hcloud_floating_ip" "ip_1" {
  ip_address = "1.2.3.4"
}
data "hcloud_floating_ip" "ip_2" {
  with_selector = "key=value"
}
resource "hcloud_floating_ip_assignment" "main" {
  count          = var.counter
  floating_ip_id = data.hcloud_floating_ip.ip_1.id
  server_id      = hcloud_server.main.id
}
```
## Argument Reference
- `id` - (Optional, string) ID of the Floating IP.
- `name` - (Optional, string) Name of the Floating IP.
- `ip_address` - (Optional, string) IP address of the Floating IP.
- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)

## Attributes Reference
- `id` - (int) Unique ID of the Floating IP.
- `type` - (string) Type of the Floating IP.
- `name` - (string) Name of the Floating IP.
- `server_id` - (int) Server to assign the Floating IP is assigned to.
- `home_location` - (string) Home location.
- `description` - (string) Description of the Floating IP.
- `ip_address` - (string) IP Address of the Floating IP.
- `ip_network` - (string) IPv6 subnet. (Only set if `type` is `ipv6`)
- `labels` - (map) User-defined labels (key-value pairs).
- `delete_protection` - (bool) Whether delete protection is enabled.
