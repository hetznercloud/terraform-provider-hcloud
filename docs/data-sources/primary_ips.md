---
page_title: "Hetzner Cloud: hcloud_primary_ips"
description: |-
  Provides details about multiple Hetzner Cloud Primary IPs.
---

# Data Source: hcloud_primary_ips

Provides details about multiple Hetzner Cloud Primary IPs.

## Example Usage

```terraform
data "hcloud_primary_ips" "ip_2" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/reference/cloud#label-selector)

## Attributes Reference

- `primary_ips` - (list) List of all matching primary ips. See `data.hcloud_primary_ip` for schema.
