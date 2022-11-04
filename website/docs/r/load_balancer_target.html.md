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
  load_balancer_id = hcloud_load_balancer.load_balancer.id
  server_id        = hcloud_server.my_server.id
}
```

## Argument Reference

- `type` - (Required, string) Type of the target. Possible values
  `server`, `label_selector`, `ip`.
- `load_balancer_id` - (Required, int) ID of the Load Balancer to which
  the target gets attached.
- `server_id` - (Optional, int) ID of the server which should be a
  target for this Load Balancer. Required if `type` is `server`
- `label_selector` - (Optional, string) Label Selector selecting targets
  for this Load Balancer. Required if `type` is `label_selector`.
- `ip` - (Optional, string) IP address for an IP Target. Required if
  `type` is `ip`.
- `use_private_ip` - (Optional, bool) use the private IP to connect to
  Load Balancer targets. Only allowed if type is `server` or
  `label_selector`.

## Attributes Reference

- `type` - (string) Type of the target. `server`
- `server_id` - (int) ID of the server which should be a target for this
  Load Balancer.
- `label_selector` - (string) Label Selector selecting targets for this
  Load Balancer.
- `ip` - (string) IP address of an IP Target.
- `use_private_ip` - (bool) use the private IP to connect to Load
  Balancer targets.

## Import

Load Balancer Target entries can be imported using a compound ID with the following format:
`<load-balancer-id>__<type>__<identifier>`

Where _identifier_ depends on the _type_:

- `server`: server id, for example: `123`
- `label_selector`: label selector, for example: `foo=bar`
- `ip`: ip address, for example: `203.0.113.123`

```
terraform import hcloud_load_balancer_target.myloadbalancerservertarget 123__server__321
terraform import hcloud_load_balancer_target.myloadbalancerlabeltarget 123__label_selector__foo=bar
terraform import hcloud_load_balancer_target.myloadbalanceriptarget 123__ip__203.0.113.123
```
