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
- `most_recent` - (Optional, bool) Return most recent firewall if multiple are found.

## Attribute Reference

- `id` - (int) Unique ID of the Firewall.
- `name` - (string) Name of the Firewall.
- `rule` - (string)  Configuration of a Rule from this Firewall.
- `labels` - (map) User-defined labels (key-value pairs)
- `apply_to` - Configuration of the Applied Resources

`rule` support the following fields:
- `direction` - (Required, string) Direction of the Firewall Rule. `in`, `out`
- `protocol` - (Required, string) Protocol of the Firewall Rule. `tcp`, `icmp`, `udp`, `gre`, `esp`
- `port` - (Required, string) Port of the Firewall Rule. Required when `protocol` is `tcp` or `udp`
- `source_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction` is `in`)
- `destination_ips` - (Required, List) List of CIDRs that are allowed within this Firewall Rule (when `direction` is `out`)
- `description` - (Optional, string) Description of the firewall rule

`apply_to` support the following fields:
- `label_selector` - (string) Label Selector to select servers the firewall is applied to. Empty if a server is directly
  referenced
- `server` - (int) ID of a server where the firewall is applied to. `0` if applied to a label_selector
