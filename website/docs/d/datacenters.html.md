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
  count = 5

  name        = "node${count.index}"
  image       = "debian-11"
  server_type = "cx31"
  datacenter  = element(data.hcloud_datacenters.ds.datacenters, count.index).name
}
```

## Attributes Reference
- `datacenter_ids` - (list) List of unique datacenter identifiers. **Deprecated**: Use `datacenters` attribute instead.
- `names` - (list) List of datacenter names. **Deprecated**: Use `datacenters` attribute instead.
- `descriptions` - (list) List of all datacenter descriptions. **Deprecated**: Use `datacenters` attribute instead.
- `datacenters` - (list) List of all datacenters. See `data.hcloud_datacenter` for schema.
