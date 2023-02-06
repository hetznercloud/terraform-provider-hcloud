---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_primary_ip"
sidebar_current: "docs-hcloud-datasource-primary-ip"
description: |-
Provides details about a specific Hetzner Cloud Primary IP.
---

# Data Source: hcloud_primary_ip

Provides details about a Hetzner Cloud Primary IP.

This resource can be useful when you need to determine a Primary IP ID based on the IP address.

Side note:

If a server is getting created, it has to have a primary ip. If a server is getting created without defining primary ips, two of them (one ipv4 and one ipv6) getting created & attached.
Currently, Primary IPs can be only attached to servers.

## Example Usage

# Data Source: hcloud_primary_ip

Provides details about a Hetzner Cloud Primary IP.
This resource can be useful when you need to determine a Primary IP ID based on the IP address.

## Example Usage

```hcl
data "hcloud_primary_ip" "ip_1" {
  ip_address = "1.2.3.4"
}
data "hcloud_primary_ip" "ip_2" {
  name = "primary_ip_1"
}
data "hcloud_primary_ip" "ip_3" {
  with_selector = "key=value"
}

// Link a server to an existing primary IP
resource "hcloud_server" "server_test" {
  name        = "test-server"
  image       = "ubuntu-20.04"
  server_type = "cx11"
  datacenter  = "fsn1-dc14"
  labels = {
    "test" : "tessst1"
  }
  public_net {
    ipv4 = hcloud_primary_ip.ip_1.id
  }

}
```

## Argument Reference
- `id` - (Optional, string) ID of the Primary IP.
- `name` - (Optional, string) Name of the Primary IP.
- `ip_address` - (Optional, string) IP address of the Primary IP.
- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)

## Attributes Reference
- `id` - (int) Unique ID of the Primary IP.
- `type` - (string) Type of the Primary IP.
- `name` - (string) Name of the Primary IP.
- `datacenter` - (string) The datacenter name of the Primary IP.
- `auto_delete` - (bool) Whether auto delete is enabled.
- `labels` - (string) Description of the Primary IP.
- `ip_address` - (string) IP Address of the Primary IP.
- `ip_network` - (string) IPv6 subnet of the Primary IP for IPv6 addresses. (Only set if `type` is `ipv6`)
- `assignee_id` - (int) ID of the assigned resource.
- `assignee_type` - (string) The type of the assigned resource.
- `delete_protection` - (bool) Whether delete protection is enabled.
