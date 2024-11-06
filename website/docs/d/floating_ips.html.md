---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_floating_ips"
sidebar_current: "docs-hcloud-datasource-floating-ips-x"
description: |-
Provides details about multiple Hetzner Cloud Floating IPs.
---

# Data Source: hcloud_floating_ips

Provides details about multiple Hetzner Cloud Floating IPs.

## Example Usage

```hcl
data "hcloud_floating_ips" "ip_2" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)

## Attributes Reference

- `floating_ips` - (list) List of all matching floating ips. See `data.hcloud_floating_ip` for schema.
