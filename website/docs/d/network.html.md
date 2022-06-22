---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_network"
sidebar_current: "docs-hcloud-datasource-network"
description: |-
  Provides details about a specific Hetzner Cloud network.
---
# Data Source: hcloud_network
Provides details about a Hetzner Cloud network.
This resource is useful if you want to use a non-terraform managed network.
## Example Usage
```hcl
data "hcloud_network" "network_1" {
  id = 1234
}
data "hcloud_network" "network_2" {
  name = "my-network"
}
data "hcloud_network" "network_3" {
  with_selector = "key=value"
}
```

## Argument Reference
- `id` - ID of the Network.
- `name` - Name of the Network.
- `with_selector` - Label Selector. For more information about possible values, visit the [Hetzner Cloud Documentation](https://docs.hetzner.cloud/#overview-label-selector).

## Attributes Reference
- `id` - Unique ID of the Network.
- `name` - Name of the Network.
- `ip_range` - IPv4 prefix of the Network.
- `delete_protection` - (boolean) Whether delete protection is enabled.
