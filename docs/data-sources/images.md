---
page_title: "Hetzner Cloud: hcloud_images"
description: |-
  Provides details about multiple Hetzner Cloud Images.
---

# Data Source: hcloud_images

Provides details about multiple Hetzner Cloud Images.

When relevant, it is recommended to always provide the image architecture
(`with_architecture`) when fetching images.

## Example Usage

```terraform
data "hcloud_images" "by_architecture" {
  with_architecture = ["x86"]
}

data "hcloud_images" "by_label" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/reference/cloud#label-selector)
- `most_recent` - (Optional, bool) Sorts list by date.
- `with_status` - (Optional, list) List only images with the specified status, could contain `creating` or `available`.
- `with_architecture` - (Optional, list) List only images with this architecture, could contain `x86` or `arm`.
- `include_deprecated` - (Optional, bool) Also list images that are marked as deprecated.

## Attributes Reference

- `images` - (list) List of all matching images. See `data.hcloud_image` for schema.
