---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_image"
sidebar_current: "docs-hcloud-datasource-image"
description: |-
  Provides details about a specific Hetzner Cloud Image.
---
 # Data Source: hcloud_sshkey
 Provides details about a Hetzner Cloud Image.
 This resource can be useful when you need to determine a Image, that is not managed by terraform.
## Example Usage
```hcl
data "hcloud_image" "image_1" {
  id = "1234"
}
data "hcloud_image" "image_2" {
  name = "ubuntu-18.04"
}

resource "hcloud_server" "main" {
  image  = "${data.hcloud_image.image_1.name}"
}
```
 ## Argument Reference
 - `id` - ID of the Image.
 - `name` - Name of the Image.
 ## Attributes Reference
 - `id` - Unique ID of the Image.
- `name` - Name of the Image.
- `type` - Type of the Image, could be `system`,`backup` or `snapshot.
- `status` - Status of the Image.
- `description` - Description of the Image.
- `created` - Date when the Image was created (in ISO-8601 format).
- `os_flavor` - Flavor of operating system contained in the image, could be `ubuntu`, `centos`, `debian`, `fedora` or `unknown`.
- `os_version` - Operating system version.
- `rapid_deploy` - Indicates that rapid deploy of the image is available.
- `deprecated` - Point in time when the image is considered to be deprecated (in ISO-8601 format).