---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_datacenters"
sidebar_current: "docs-hcloud-datasource-datacenters"
description: |-
  List all available Hetzner Cloud Datacenters.
---
# Data Source: hcloud_datacenters
Provides a list of available Hetzner Cloud Datacenters.
This resource may be useful to create highly available infrastructure, distributed across several datacenters.

## Example Usage
```hcl
data "hcloud_datacenters" "ds" {
}

resource "hcloud_server" "workers" {
  count = 3

  name        = "node1"
  image       = "debian-11"
  server_type = "cx31"
  datacenter  = element(data.hcloud_datacenters.ds.names, count.index)
}
```

## Attributes Reference
- `datacenter_ids` - (list) List of unique datacenter identifiers.
- `names` - (list) List of datacenter names.
- `descriptions` - (list) List of all datacenter descriptions.
- `datacenters` - (list) List of all datacenters. See `data.hcloud_datacenter` for schema.
