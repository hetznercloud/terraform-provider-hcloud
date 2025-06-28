---
page_title: "Hetzner Cloud: hcloud_iso"
description: |-
  Provides details about a specific Hetzner Cloud ISO.
---

# Data Source: hcloud_iso

Provides details about a Hetzner Cloud ISO.

When relevant, it is recommended to always provide the targeted architecture
(`with_architecture`) when fetching ISOs.

## Example Usage

```terraform
data "hcloud_iso" "by_id" {
  id = "117577"
}

data "hcloud_iso" "by_name_x86" {
  name              = "nixos-minimal-24.11.712431.cbd8ec4de446-x86_64-linux.iso"
  with_architecture = "x86"
}

data "hcloud_iso" "by_name_prefix_arm" {
  name_prefix       = "nixos-minimal-24.11"
  with_architecture = "arm"
}

resource "hcloud_server" "main" {
  iso = data.hcloud_iso.by_name_x86.id
}
```

## Argument Reference

- `id` - (Optional, string) ID of the ISO.
- `name` - (Optional, string) Name of the ISO.
- `name_prefix` - (Optional, string) List only ISOs with names starting with this prefix.
- `type` - (Optional, string) List only ISOs with the associated type, could be `public` or `private`.
- `with_architecture` - (Optional, string) Select only ISOs with this architecture, could be `x86` (default) or `arm`.
- `include_architecture_wildcard` - (Optional, boolean) If set to `true`, return custom ISOs that have no architecture set

## Attributes Reference

- `id` - (int) Unique ID of the ISO.
- `name` - (string) Name of the ISO.
- `type` - (string) Type of the ISO, could be `public` or `private`.
- `architecture` - (string) If defined, architecture of the ISO, could be `x86` or `arm`
- `description` - (string) Description of the ISO.
- `deprecated_announced` - (string) If defined, date when the ISO will be deprecated (in ISO-8601 format).
- `unavailable_after` - (string) If defined, date when the ISO will be removed (in ISO-8601 format).