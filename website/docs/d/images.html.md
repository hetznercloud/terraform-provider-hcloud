---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_images"
sidebar_current: "docs-hcloud-datasource-images-x"
description: |-
Provides details about multiple Hetzner Cloud Images.
---

# Data Source: hcloud_images
Provides details about multiple Hetzner Cloud Images.


## Example Usage
```hcl
data "hcloud_images" "image_2" {
}

data "hcloud_images" "image_3" {
  with_selector = "key=value"
}
```
## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)
- `most_recent` - (Optional, bool) Sorts list by date.
- `with_status` - (Optional, list) List only images with the specified status, could contain `creating` or `available`.

## Attributes Reference
- `images` - (list) List of all matching images. See `data.hcloud_image` for schema.
