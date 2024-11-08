---
page_title: "Hetzner Cloud: hcloud_server_types"
description: |-
  List all available Hetzner Cloud Server Types.
---

# Data Source: hcloud_server_types

Provides a list of available Hetzner Cloud Server Types.

## Example Usage

```terraform
data "hcloud_server_types" "all" {}
```

## Attributes Reference

- `server_types` - (list) List of all server types. See `data.hcloud_server_type` for schema.
