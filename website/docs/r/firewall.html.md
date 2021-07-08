---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_firewall"
sidebar_current: "docs-hcloud-resource-firewall"
description: |-
Provides a Hetzner Cloud Firewall to represent a Firewall in the Hetzner Cloud.
---

# hcloud_firewall

Provides a Hetzner Cloud Firewall to represent a Firewall in the Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_firewall" "myfirewall" {
  name = "my-firewall"
  rule {
   direction = "in"
   protocol  = "icmp"
   source_ips = [
      "0.0.0.0/0",
      "::/0"
   ]
  }
  apply_to {
    server = 22
  }
  apply_to {
    labelSelector = "security=production"
  }
}

resource "hcloud_server" "node1" {
  name         = "node1"
  image        = "debian-9"
  server_type  = "cx11"
  firewall_ids = [hcloud_firewall.myfirewall.id]
}
```

## Argument Reference

- `name` - (Optional, string) Name of the Firewall.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.
- `rule` - (Optional) Configuration of a Rule from this Firewall.
- `apply_to` - (Optional) Configuration of an ApplyTo from this Firewall.

`rule` support the following fields:
- `direction` - (Required, string) Direction of the Firewall Rule. `in`
- `protocol` - (Required, string) Protocol of the Firewall Rule. `tcp`, `icmp`, `udp`, `gre`, `esp`
- `port` - (Required, string) Port of the Firewall Rule. Required when `protocol` is `tcp` or `udp`. You can use `any` to allow all ports for the specific protocol. Port ranges are also possible: `80:85` allows all ports between 80 and 85.
- `source_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule

`apply_to` support the following fields:
- `server` - (Optional, int) ID of the server to apply firewall to
- `labelSelectors` - (Optional, string) Label selector of the firewall

## Attributes Reference

- `id` - (int) Unique ID of the Firewall.
- `name` - (string) Name of the Firewall.
- `rule` - (string)  Configuration of a Rule from this Firewall.
- `labels` - (map) User-defined labels (key-value pairs)
- `apply_to` - (map) Configuration of an ApplyTo from this Firewall.

`rule` support the following fields:
- `direction` - (Required, string) Direction of the Firewall Rule. `in`, `out`
- `protocol` - (Required, string) Protocol of the Firewall Rule. `tcp`, `icmp`, `udp`, `gre`, `esp`
- `port` - (Required, string) Port of the Firewall Rule. Required when `protocol` is `tcp` or `udp`
- `source_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction` is `in`)
- `destination_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction` is `out`)

`apply_to` support the following fields:
- `server` - (Optional, int) ID of the server to apply firewall to
- `labelSelectors` - (Optional, string) Label selector of the firewall

## Import

Firewalls can be imported using its `id`:

```
terraform import hcloud_firewall.myfw <id>
```
