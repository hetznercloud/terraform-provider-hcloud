---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_primary_ips"
sidebar_current: "docs-hcloud-datasource-primary-ips-x"
description: |-
Provides details about multiple Hetzner Cloud Primary IPs.
---

# Data Source: hcloud_primary_ips
Provides details about multiple Hetzner Cloud Primary IPs.


## Example Usage
```hcl
data "hcloud_primary_ips" "ip_2" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)

## Attributes Reference
- `primary_ips` - (list) List of all matching primary ips. See `data.hcloud_primary_ip` for schema.
