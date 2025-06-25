---
page_title: "Hetzner Cloud: hcloud_load_firewalls"
description: |-
  Provides details about multiple Hetzner Cloud Firewall.
---

# hcloud_firewalls

Provides details about multiple Hetzner Cloud Firewall.

## Example Usage

```terraform
data "hcloud_firewalls" "sample_firewall_1" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/reference/cloud#label-selector)
- `most_recent` - (Optional, bool) Sorts list by date.

## Attribute Reference

- `firewalls` - (list) List of all matching firewalls. See `data.hcloud_firewall` for schema.
