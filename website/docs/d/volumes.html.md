---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_volume"
sidebar_current: "docs-hcloud-datasource-volumes-x"
description: |-
Provides details about multiple Hetzner Cloud volumes.
---

# hcloud_volumes
Provides details about multiple Hetzner Cloud volumes.


## Example Usage
```hcl
data "hcloud_volumes" "volume_" {

}
data "hcloud_volumes" "volume_3" {
  with_selector = "key=value"
}
```

## Argument Reference
- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)
- `with_status` - (Optional, list) List only volumes with the specified status, could contain `creating` or `available`.

## Attributes Reference
- `volumes` - (list) List of all matching volumes. See `data.hcloud_volume` for schema.
