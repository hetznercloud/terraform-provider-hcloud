---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_volume"
sidebar_current: "docs-hcloud-resource-volume-x"
description: |-
  Provides a Hetzner Cloud volume resource to manage volumes.
---

# hcloud_volume

Provides a Hetzner Cloud volume resource to manage volumes.

## Example Usage

```hcl
resource "hcloud_server" "node1" {
  name = "node1"
  image = "debian-9"
  server_type = "cx11"
}

resource "hcloud_volume" "master" {
  name = "volume1"
  size = 50
  server_id = "${hcloud_server.node1.id}"
}
```

## Argument Reference

- `name` - (Required, string) Name of the volume to create (must be unique per project).
- `size` - (Required, int) Size of the volume (in GB).
- `server` - (Optional, int) Server to attach the Volume to, optional if location argument is passed.
- `location` - (Optional, string) Location of the volume to create, optional if server_id argument is passed.


## Attributes Reference

- `id` - Unique ID of the volume.
- `name` - Name of the volume.
- `size` - Size of the volume.
- `labels` - User-defined labels (key-value pairs).
- `linux_device` - 	Device path on the file system for the Volume.


## Import

Volumes can be imported using their `id`:

```
terraform import hcloud_volume.myvolume <id>
```