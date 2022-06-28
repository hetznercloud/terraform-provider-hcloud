---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_networks"
sidebar_current: "docs-hcloud-datasource-networks-x"
description: |-
Provides details about multiple Hetzner Cloud Networks.
---

# hcloud_networks
Provides details about multiple Hetzner Cloud Networks.


## Example Usage
```hcl
data "hcloud_network" "network_2" {

}
data "hcloud_network" "network_3" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)

## Attributes Reference
- `networks` - (list) List of all matching networks. See `data.hcloud_network` for schema.
