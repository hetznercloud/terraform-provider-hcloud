---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_load_firewall"
sidebar_current: "docs-hcloud-datasource-firewall-x"
description: |-
Provides details about a specific Hetzner Cloud Firewall.
---

# hcloud_firewall

Provides details about a specific Hetzner Cloud Firewall.

```hcl
data "hcloud_firewall" "sample_firewall_1" {
  name = "sample-firewall-1"
}

data "hcloud_firewall" "sample_firewall_2" {
  id = "4711"
}
```

## Argument Reference

- `id` - ID of the firewall.
- `name` - Name of the firewall.
- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)

## Attribute Reference

- `id` - (int) Unique ID of the Firewall.
- `name` - (string) Name of the Firewall.
- `rule` - (string)  Configuration of a Rule from this Firewall.
- `labels` - (map) User-defined labels (key-value pairs)

`rule` support the following fields:
- `direction` - (Required, string) Direction of the Firewall Rule. `in`, `out`
- `protocol` - (Required, string) Protocol of the Firewall Rule. `tcp`, `icmp`, `udp`
- `port` - (Required, string) Port of the Firewall Rule. Required when `protocol` is `tcp` or `udp`
- `source_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction` is `in`)
- `destination_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction` is `out`)
