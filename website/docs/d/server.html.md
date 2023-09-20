---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_server"
sidebar_current: "docs-hcloud-datasource-server-x"
description: |-
  Provides details about a specific Hetzner Cloud Server.
---
# Data Source: hcloud_server
Provides details about a Hetzner Cloud Server.
This resource is useful if you want to use a non-terraform managed server.

## Example Usage
```hcl
data "hcloud_server" "s_1" {
  name = "my-server"
}
data "hcloud_server" "s_2" {
  id = "123"
}
data "hcloud_server" "s_3" {
  with_selector = "key=value"
}
```

## Argument Reference
- `id` - ID of the server.
- `name` - Name of the server.
- `with_selector` - Label Selector. For more information about possible values, visit the [Hetzner Cloud Documentation](https://docs.hetzner.cloud/#overview-label-selector).
- `with_status` - (Optional, list) List only servers with the specified status, could contain `initializing`, `starting`, `running`, `stopping`, `off`, `deleting`, `rebuilding`, `migrating`, `unknown`.

## Attributes Reference
- `id` - (int) Unique ID of the server.
- `name` - (string) Name of the server.
- `server_type` - (string) Name of the server type.
- `image` - (string) Name or ID of the image the server was created from.
- `location` - (string) The location name.
- `datacenter` - (string) The datacenter name.
- `backup_window` - (string) The backup window of the server, if enabled.
- `backups` - (bool) Whether backups are enabled.
- `iso` - (string) ID or Name of the mounted ISO image. Architecture of ISO must equal the server (type) architecture.
- `ipv4_address` - (string) The IPv4 address.
- `ipv6_address` - (string) The first IPv6 address of the assigned network.
- `ipv6_network` - (string) The IPv6 network.
- `status` - (string) The status of the server.
- `labels` - (map) User-defined labels (key-value pairs)
- `firewall_ids` - (Optional, list) Firewall IDs the server is attached to.
- `placement_group_id` - (Optional, string) Placement Group ID the server is assigned to.
- `delete_protection` - (bool) Whether delete protection is enabled.
- `rebuild_protection` - (bool) Whether rebuild protection is enabled.
