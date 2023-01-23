---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_locations"
sidebar_current: "docs-hcloud-datasource-locations"
description: |-
  List all available Hetzner Cloud Locations.
---
# Data Source: hcloud_locations
Provides a list of available Hetzner Cloud Locations.
This resource may be useful to create highly available infrastructure, distributed across several locations.

## Example Usage
```hcl
data "hcloud_locations" "ds" {
}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-11"
  server_type = "cx31"
  location    = element(data.hcloud_locations.ds.locations, count.index).name
}
```

## Attributes Reference
- `location_ids` - (list) List of unique location identifiers. **Deprecated**: Use `locations` attribute instead.
- `names` - (list) List of location names. **Deprecated**: Use `locations` attribute instead.
- `descriptions` - (list) List of all location descriptions. **Deprecated**: Use `locations` attribute instead.
- `locations` - (list) List of all locations. See `data.hcloud_location` for schema.
