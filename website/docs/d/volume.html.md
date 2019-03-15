---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_volume"
sidebar_current: "docs-hcloud-datasource-volume"
description: |-
  Provides details about a specific Hetzner Cloud volume.
---
# Data Source: hcloud_volume
Provides details about a Hetzner Cloud volume.
This resource is useful if you want to use a non-terraform managed volume.
## Example Usage
```hcl
data "hcloud_volume" "volume_1" {
  id = "1234"
}
data "hcloud_volume" "volume_2" {
  name = "my-volume"
}
data "hcloud_volume" "volume_3" {
  selector = "key=value"
}
```

## Argument Reference
- `id` - ID of the volume.
- `name` - Name of the volume.
- `selector` - Label Selector. For more information about possible values, visit the [Hetzner Cloud Documentation](https://docs.hetzner.cloud/#overview-label-selector).
- `with_status` - (Optional, list) List only volumes with the specified status, could contain `creating` or `available`.

## Attributes Reference
- `id` - Unique ID of the volume.
- `name` - Name of the volume.
- `size` - Size of the volume.