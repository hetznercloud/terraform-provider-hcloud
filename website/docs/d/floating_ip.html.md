---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_floating_ip"
sidebar_current: "docs-hcloud-datasource-floating-ip"
description: |-
  Provides details about a specific Hetzner Cloud Floating IP.
---
<<<<<<< HEAD

# Data Source: hcloud_floating_ip

Provides details about a Hetzner Cloud Floating IP.

This resource can be useful when you need to determine a Floating IP ID based on the IP address.

## Example Usage

```hcl
data "hcloud_floating_ip" "ip_1" {
  ip_address = "1.2.3.4"
}

resource "hcloud_floating_ip_association" "main" {
=======
 # Data Source: hcloud_floating_ip
 Provides details about a Hetzner Cloud Floating IP.
 This resource can be useful when you need to determine a Floating IP ID based on the IP address.
 ## Example Usage
 ```hcl
data "hcloud_floating_ip" "ip_1" {
  ip_address = "1.2.3.4"
}
 resource "hcloud_floating_ip_association" "main" {
>>>>>>> 6dbdb17... Add tests for floating ip datasources
  count          = "${var.counter}"
  floating_ip_id = "${data.hcloud_floating_ip.ip_1.id}"
  server_id      = "${hcloud_server.main.id}"
}
```
<<<<<<< HEAD

## Argument Reference

- `ip_address` - IP address of the Floating IP.

## Attributes Reference

- `id` - Unique ID of the Floating IP.
- `ip_address` - IP address of the Floating IP.
=======
 ## Argument Reference
 - `ip_address` - IP address of the Floating IP.
 ## Attributes Reference
 - `id` - Unique ID of the Floating IP.
- `ip_address` - IP address of the Floating IP.
>>>>>>> 6dbdb17... Add tests for floating ip datasources
