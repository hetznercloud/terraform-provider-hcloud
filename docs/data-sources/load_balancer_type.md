---
page_title: "Hetzner Cloud: hcloud_load_balancer_type"
description: |-
  Provides details about a specific Hetzner Cloud Load Balancer Type.
---

# Data Source: hcloud_load_balancer_type

Provides details about a specific Hetzner Cloud Load Balancer Type.
Use this resource to get detailed information about specific Load Balancer Type.

## Example Usage

```terraform
data "hcloud_load_balancer_type" "by_name" {
  name = "cx22"
}

data "hcloud_load_balancer_type" "by_id" {
  id = 1
}

resource "hcloud_load_balancer" "load_balancer" {
  name               = "my-load-balancer"
  load_balancer_type = data.hcloud_load_balancer_type.name
  location           = "nbg1"
}
```

## Argument Reference

- `id` - (Optional, string) ID of the load_balancer_type.
- `name` - (Optional, string) Name of the load_balancer_type.

## Attributes Reference

- `id` - (int) Unique ID of the load_balancer_type.
- `name` - (string) Name of the load_balancer_type.
- `description` - (string) Description of the load_balancer_type.
- `max_assigned_certificates` - (int) Maximum number of SSL Certificates that can be assigned to the Load Balancer of this type.
- `max_connections` - (int) Maximum number of simultaneous open connections for the Load Balancer of this type.
- `max_services` - (int) Maximum number of services for the Load Balancer of this type.
- `max_targets` - (int) Maximum number of targets for the Load Balancer of this type.
