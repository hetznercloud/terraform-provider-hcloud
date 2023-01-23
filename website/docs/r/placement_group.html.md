---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_placement_group"
sidebar_current: "docs-hcloud-placement-group"
description: |-
Provides a Hetzner Cloud Placement Group to represent a Placement Group in the Hetzner Cloud.
---

# hcloud_placement_group

Provides a Hetzner Cloud Placement Group to represent a Placement Group in the Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_placement_group" "my-placement-group" {
  name = "my-placement-group"
  type = "spread"
  labels = {
    key = "value"
  }
}

resource "hcloud_server" "node1" {
  name         = "node1"
  image        = "debian-11"
  server_type  = "cx11"
  placement_group_id = hcloud_placement_group.my-placement-group.id
}
```

## Argument Reference

- `name` - (Optional, string) Name of the Placement Group.
- `type` - (Required, string) Type of the Placement Group.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.

## Attributes Reference

- `id` - (int) Unique ID of the Placement Group.
- `name` - (string) Name of the Placement Group.
- `type` - (string)  Type of the Placement Group.
- `labels` - (map) User-defined labels (key-value pairs)

## Import

Placement Groups can be imported using its `id`:

```
terraform import hcloud_placement_group.my-placement-group id
```
