---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_location"
sidebar_current: "docs-hcloud-datasource-location"
description: |-
  Provides details about a specific Hetzner Cloud Location.
---
# Data Source: hcloud_location
Provides details about a specific Hetzner Cloud Location.
Use this resource to get detailed information about specific location.

## Example Usage
```hcl
data "hcloud_location" "l_1" {
  name = "fsn1"
}
data "hcloud_location" "l_2" {
  id = 1
}
```
## Argument Reference
- `id` - (Optional, string) ID of the location.
- `name` - (Optional, string) Name of the location.

## Attributes Reference
- `id` - (int) Unique ID of the location.
- `name` - (string) Name of the location.
- `description` - (string) Description of the location.
- `city` - (string) City of the location.
- `country` - (string) Country of the location.
- `latitude` - (float) Latitude of the city.
- `longitude` - (float) Longitude of the city.
