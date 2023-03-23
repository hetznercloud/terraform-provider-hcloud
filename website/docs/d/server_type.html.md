---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_server_type"
sidebar_current: "docs-hcloud-datasource-server-type"
description: |-
  Provides details about a specific Hetzner Cloud Server Type.
---
# Data Source: hcloud_server_type
Provides details about a specific Hetzner Cloud Server Type.
Use this resource to get detailed information about specific Server Type.

## Example Usage
```hcl
data "hcloud_server_type" "ds_1" {
  name = "cx11"
}
data "hcloud_server_type" "ds_2" {
  id = 1
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
