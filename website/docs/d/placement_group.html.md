---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_placement_group"
sidebar_current: "docs-hcloud-datasource-placement-group"
description: |-
  Provides details about a specific Hetzner Cloud Placement Group.
---

# hcloud_placement_group

Provides details about a specific Hetzner Cloud Placement Group.

```hcl
data "hcloud_placement_group" "sample_placement_group_1" {
  name = "sample-placement-group-1"
}

data "hcloud_placement_group" "sample_placement_group_2" {
  id = "4711"
}
```

## Argument Reference

- `id` - ID of the placement group.
- `name` - Name of the placement group.
- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)
- `most_recent` - (Optional, bool) Return most recent placement group if multiple are found.

## Attribute Reference

- `id` - (int) Unique ID of the Placement Group.
- `name` - (string) Name of the Placement Group.
- `type` - (string) Type of the Placement Group.
- `labels` - (map) User-defined labels (key-value pairs)
