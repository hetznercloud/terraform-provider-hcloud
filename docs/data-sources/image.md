---
page_title: "Hetzner Cloud: hcloud_image"
description: |-
  Provides details about a specific Hetzner Cloud Image.
---

# Data Source: hcloud_image

Provides details about a Hetzner Cloud Image.
This resource is useful if you want to use a non-terraform managed image.

When relevant, it is recommended to always provide the image architecture
(`with_architecture`) when fetching images.

## Example Usage

```terraform
data "hcloud_image" "by_id" {
  id = "114690387"
}

data "hcloud_image" "by_name_x86" {
  name              = "debian-12"
  with_architecture = "x86"
}

data "hcloud_image" "by_name_arm" {
  name              = "debian-12"
  with_architecture = "arm"
}

data "hcloud_image" "by_label" {
  with_selector = "key=value"
}

resource "hcloud_server" "main" {
  image = data.hcloud_image.by_name.id
}
```

## Argument Reference

- `id` - (Optional, string) ID of the Image.
- `name` - (Optional, string) Name of the Image.
- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)
- `most_recent` - (Optional, bool) If more than one result is returned, use the most recent Image.
- `with_status` - (Optional, list) Select only images with the specified status, could contain `creating` or `available`.
- `with_architecture` - (Optional, string) Select only images with this architecture, could be `x86` (default) or `arm`.
- `include_deprecated` - (Optional, bool) Also return the image if it is marked as deprecated.

## Attributes Reference

- `id` - (int) Unique ID of the Image.
- `name` - (string) Name of the Image, only present when the Image is of type `system`.
- `type` - (string) Type of the Image, could be `system`, `backup` or `snapshot`.
- `status` - (string) Status of the Image.
- `description` - (string) Description of the Image.
- `created` - (string) Date when the Image was created (in ISO-8601 format).
- `os_flavor` - (string) Flavor of operating system contained in the image, could be `ubuntu`, `centos`, `debian`, `fedora` or `unknown`.
- `os_version` - (string) Operating system version.
- `rapid_deploy` - (bool) Indicates that rapid deploy of the image is available.
- `deprecated` - (string) Point in time when the image is considered to be deprecated (in ISO-8601 format).
- `architecture` - (string) Architecture of the Image.
