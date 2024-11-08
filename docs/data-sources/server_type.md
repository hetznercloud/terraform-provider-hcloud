---
page_title: "Hetzner Cloud: hcloud_server_type"
description: |-
  Provides details about a specific Hetzner Cloud Server Type.
---

# Data Source: hcloud_server_type

Provides details about a specific Hetzner Cloud Server Type.
Use this resource to get detailed information about specific Server Type.

## Example Usage

```terraform
data "hcloud_server_type" "by_id" {
  id = 22
}

data "hcloud_server_type" "by_name" {
  name = "cx22"
}

resource "hcloud_server" "main" {
  name        = "my-server"
  location    = "fsn1"
  image       = "debian-12"
  server_type = data.hcloud_server_type.by_name.name
}
```

## Argument Reference

- `id` - (Optional, string) ID of the server_type.
- `name` - (Optional, string) Name of the server_type.

## Attributes Reference

- `id` - (int) Unique ID of the server_type.
- `name` - (string) Name of the server_type.
- `description` - (string) Description of the server_type.
- `cores` - (int) Number of cpu cores a Server of this type will have.
- `memory` - (int) Memory a Server of this type will have in GB.
- `disk` - (int) Disk size a Server of this type will have in GB.
- `architecture` - (string) Architecture of the server_type.
- `included_traffic` - (int) Free traffic per month in bytes. **Warning**: This field is deprecated and will report `0` after 2024-08-05.
- `is_deprecated` - (bool) Deprecation status of server type.
- `deprecation_announced` (Optional, string) Date when the deprecation of the server type was announced. Only set when the server type is deprecated.
- `unavailable_after` (Optional, string) Date when the server type will not be available for new servers. Only set when the server type is deprecated.