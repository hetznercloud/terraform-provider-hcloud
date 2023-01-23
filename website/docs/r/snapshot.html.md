---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_snapshot"
sidebar_current: "docs-hcloud-resource-snapshot"
description: |-
  Provides a Hetzner Cloud snapshot to represent an image with type snapshot in the Hetzner Cloud.
---

# hcloud_snapshot

Provides a Hetzner Cloud snapshot to represent an image with type snapshot in the Hetzner Cloud. This resource makes it easy to create a snapshot of your server.

## Example Usage

```hcl
resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-11"
  server_type = "cx11"
}

resource "hcloud_snapshot" "my-snapshot" {
  server_id = hcloud_server.node1.id
}
```

## Argument Reference

- `server_id` - (Required, int) Server to the snapshot should be created from.
- `description` - (Optional, string) Description of the snapshot.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.

## Attributes Reference

- `id` - (int) Unique ID of the snapshot.
- `server_id` - (int) Server the snapshot was created from.
- `description` - (string) Description of the snapshot.
- `labels` - (map) User-defined labels (key-value pairs)

## Import

Snapshots can be imported using its image `id`:

```
terraform import hcloud_snapshot.myimage id
```
