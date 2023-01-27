---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_firewall"
sidebar_current: "docs-hcloud-resource-firewall"
description: |- Provides a Hetzner Cloud Firewall to represent a Firewall in the Hetzner Cloud.
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

  rule {
    direction = "in"
    protocol  = "tcp"
    port      = "80-85"
    source_ips = [
      "0.0.0.0/0",
      "::/0"
    ]
  }

}

resource "hcloud_server" "node1" {
  name         = "node1"
  image        = "debian-11"
  server_type  = "cx11"
  firewall_ids = [hcloud_firewall.myfirewall.id]
}
```

## Argument Reference

- `name` - (Optional, string) Name of the Firewall.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.
- `rule` - (Optional) Configuration of a Rule from this Firewall.
- `apply_to` (Optional) Resources the firewall should be assigned to

`rule` support the following fields:

- `direction` - (Required, string) Direction of the Firewall Rule. `in`
- `protocol` - (Required, string) Protocol of the Firewall Rule. `tcp`, `icmp`, `udp`, `gre`, `esp`
- `port` - (Required, string) Port of the Firewall Rule. Required when `protocol` is `tcp` or `udp`. You can use `any`
  to allow all ports for the specific protocol. Port ranges are also possible: `80-85` allows all ports between 80 and
  85.
- `source_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule
- `description` - (Optional, string) Description of the firewall rule

`apply_to` support the following fields:

- `label_selector` - (Optional, string) Label Selector to select servers the firewall should be applied to (only one
  of `server` and `label_selector`can be applied in one block)
- `server` - (Optional, int) ID of the server you want to apply the firewall to (only one of `server`
  and `label_selector`can be applied in one block)

## Attributes Reference

- `id` - (int) Unique ID of the Firewall.
- `name` - (string) Name of the Firewall.
- `rule` - Configuration of a Rule from this Firewall.
- `labels` - (map) User-defined labels (key-value pairs)
- `apply_to` - Configuration of the Applied Resources

`rule` support the following fields:
- `direction` - (Required, string) Direction of the Firewall Rule. `in`, `out`
- `protocol` - (Required, string) Protocol of the Firewall Rule. `tcp`, `icmp`, `udp`, `gre`, `esp`
- `port` - (Required, string) Port of the Firewall Rule. Required when `protocol` is `tcp` or `udp`
- `source_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction` is `in`)
- `destination_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction`
  is `out`)
- `description` - (Optional, string) Description of the firewall rule

`apply_to` support the following fields:
- `label_selector` - (string) Label Selector to select servers the firewall is applied to. Empty if a server is directly
  referenced
- `server` - (int) ID of a server where the firewall is applied to. `0` if applied to a label_selector

## Import

Firewalls can be imported using its `id`:

```
terraform import hcloud_firewall.myfirewall id
```
