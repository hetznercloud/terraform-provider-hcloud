---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_servers"
sidebar_current: "docs-hcloud-datasource-servers-x"
description: |-
  Provides details about multiple Hetzner Cloud Servers.
---

# Data Source: hcloud_servers

Provides details about multiple Hetzner Cloud Servers.
This resource is useful if you want to use non-terraform managed servers.

## Example Usage

```hcl
data "hcloud_servers" "s_3" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) Label Selector. For more information about possible values, visit the [Hetzner Cloud Documentation](https://docs.hetzner.cloud/#overview-label-selector).
- `with_status` - (Optional, list) List only servers with the specified status, could contain `initializing`, `starting`, `running`, `stopping`, `off`, `deleting`, `rebuilding`, `migrating`, `unknown`.

## Attributes Reference

- `servers` - (list) List of all matching servers. See `data.hcloud_server` for schema.
