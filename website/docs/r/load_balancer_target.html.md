---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_load_balancer_target"
sidebar_current: "docs-hcloud-resource-load-balancer-target-x"
description: |-
  Adds a target to a Hetzner Cloud Load Balancer.
---

# hcloud_load_balancer_target

Adds a target to a Hetzner Cloud Load Balancer.

## Example Usage

```hcl
resource "hcloud_server" "my_server" {
  name        = "my-server"
  server_type = "cx11"
  image       = "ubuntu-18.04"
}

resource "hcloud_load_balancer" "load_balancer" {
  name               = "my-load-balancer"
  load_balancer_type = "lb11"
  location           = "nbg1"
}

resource "hcloud_load_balancer_target" "load_balancer_target" {
  type             = "server"
  load_balancer_id = "${hcloud_load_balancer.load_balcancer.id}"
  server_id        = "${hcloud_server.my_server.id}"
}
```

## Argument Reference

- `type` - (Required, string) Type of the target. `server`
- `load_balancer_id` - (Required, int) ID of the Load Balancer to which
  the target gets attached.
- `server_id` - (Optional, int) ID of the server which should be a
  target for this Load Balancer. Required if `type` is `server`
- `use_private_ip` - (Optional, string) use the private IP to connect to
  Load Balancer targets.

## Attributes Reference

- `type` - (string) Type of the target. `server`
- `server_id` - (int) ID of the server which should be a target for this
  Load Balancer.
- `use_private_ip` - (string) use the private IP to connect to Load
  Balancer targets.
