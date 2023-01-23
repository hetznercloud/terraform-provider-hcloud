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
  name        = "node1"
  image       = "debian-11"
  server_type = "cx11"
}

resource "hcloud_volume" "master" {
  name      = "volume1"
  size      = 50
  server_id = hcloud_server.node1.id
  automount = true
  format    = "ext4"
}
```

## Argument Reference

- `name` - (Required, string) Name of the volume to create (must be unique per project).
- `size` - (Required, int) Size of the volume (in GB).
- `server_id` - (Optional, int) Server to attach the Volume to, not allowed if location argument is passed.
- `location` - (Optional, string) The location name of the volume to create, not allowed if server_id argument is passed.
- `automount` - (Optional, bool) Automount the volume upon attaching it (server_id must be provided).
- `format` - (Optional, string) Format volume after creation. `xfs` or `ext4`
- `delete_protection` - (Optional, bool) Enable or disable delete protection.

**Note:** When you want to attach multiple volumes to a server, please use the `hcloud_volume_attachment` resource and the `location` argument instead of the `server_id` argument.

## Attributes Reference

- `id` - (int) Unique ID of the volume.
- `name` - (string) Name of the volume.
- `size` - (int) Size of the volume.
- `location` - (string) The location name.
- `server_id` - (Optional, int) Server ID the volume is attached to
- `labels` - (map) User-defined labels (key-value pairs).
- `linux_device` - (string) Device path on the file system for the Volume.
- `delete_protection` - (bool) Whether delete protection is enabled.

## Import

Volumes can be imported using their `id`:

```
terraform import hcloud_volume.myvolume id
```
