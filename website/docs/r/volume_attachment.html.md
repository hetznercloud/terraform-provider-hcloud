---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_volume_attachment"
sidebar_current: "docs-hcloud-resource-volume-attachment"
description: |-
  Provides a Hetzner Cloud Volume attachment to attach a Volume to a Hetzner Cloud Server.
---

# hcloud_volume_attachment

Provides a Hetzner Cloud Volume attachment to attach a Volume to a Hetzner Cloud Server. Deleting a Volume Attachment will detach the Volume from the Server.

## Example Usage

```hcl
resource "hcloud_volume_attachment" "main" {
  volume_id = hcloud_volume.master.id
  server_id = hcloud_server.node1.id
  automount = true
}

resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx11"
  datacenter  = "nbg1-dc3"
}

resource "hcloud_volume" "master" {
  location = "nbg1"
  size     = 10
}
```

## Argument Reference

- `volume_id` - (Required, int) ID of the Volume.
- `server_id` - (Required, int) Server to attach the Volume to.
- `automount` - (Optional, bool) Automount the volume upon attaching it.

## Attributes Reference

- `id` - (int) Unique ID of the Volume Attachment.
- `volume_id` - (int) ID of the Volume.
- `server_id` - (int) Server the Volume was attached to.

## Import

Volume Attachments can be imported using the `volume_id`:

```
terraform import hcloud_volume_attachment.myvolumeattachment <volume_id>
```
