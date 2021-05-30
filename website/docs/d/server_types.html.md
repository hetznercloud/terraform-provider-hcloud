---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_server_types"
sidebar_current: "docs-hcloud-datasource-server-types"
description: |-
  List all available Hetzner Cloud Server Types.
---
# Data Source: hcloud_server_types
Provides a list of available Hetzner Cloud Server Types.

## Example Usage
```hcl
data "hcloud_server_types" "ds" {
}

resource "hcloud_server" "workers" {
  count = 3

  name        = "node1"
  image       = "debian-9"
  server_type = element(data.hcloud_server_types.ds.names, count.index)
}
```

## Attributes Reference
- `server_type_ids` - (list) List of unique Server Types identifiers.
- `names` - (list) List of Server Types names.
- `descriptions` - (list) List of all Server Types descriptions.
