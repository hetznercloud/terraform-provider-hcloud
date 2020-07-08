---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_load_balancer"
sidebar_current: "docs-hcloud-resource-load-balancer-x"
description: |-
  Provides a Hetzner Cloud Load Balancer to represent a Load Balancer in the Hetzner Cloud.
---

# hcloud_load_balancer

  Provides a Hetzner Cloud Load Balancer to represent a Load Balancer in the Hetzner Cloud.

## Example Usage

```hcl
resource "hcloud_server" "myserver" {
  name       = "server-%d"
  server_type = "cx11"
  image = "ubuntu-18.04"
}

resource "hcloud_load_balancer" "load_balancer" {
  name       = "my-load-balancer"
  load_balancer_type = "lb11"
  location   = "nbg1"
  target {
    type = "server"
    server_id = "${hcloud_server.myserver.id}"
  }
}
```

## Argument Reference

- `name` - (Required, string) Name of the Load Balancer.
- `load_balancer_type` - (Required, string) Type of the Load Balancer.
- `location` - (Optional, string) Location of the Load Balancer. Require when no network_zone is set.
- `network_zone` - (Optional, string) Network Zone of the Load Balancer. Require when no location is set.
- `algorithm` - (Optional) Configuration of the algorithm the Load Balancer use.
- `target` - (Optional, list) List of targets of the Load Balancer.
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.

`algorithm` support the following fields:
- `type` - (Required, string) Type of the Load Balancer Algorithm. `round_robin` or `least_connections`

`target` support the following fields:
- `type` - (Required, string) Type of the target. `server`
- `server_id` - (Optional, int) ID of the server which should be a target for this Load Balancer. Required if `type` is `server`


## Attributes Reference

- `id` - (int) Unique ID of the Load Balancer.
- `load_balancer_type` - (string) Name of the Type of the Load Balancer.
- `name` - (string) Name of the Load Balancer.
- `location` - (string) Name of the location the Load Balancer is in.
- `ipv4` - (string) IPv4 Address of the Load Balancer.
- `ipv6` - (string) IPv4 Address of the Load Balancer.
- `algorithm` - (Optional) Configuration of the algorithm the Load Balancer use.
- `target` - (list) List of targets of the Load Balancer.
- `service` - (list) List of services a Load Balancer provides.
- `labels` - (map) User-defined labels (key-value pairs) .

`algorithm` support the following fields:
- `type` - (string) Type of the Load Balancer Algorithm. `round_robin` or `least_connections`

`target` supports the following fields, which are a restricted sub-set
of the fields supported by `hcloud_load_balancer_target`:
- `type` - (string) Type of the target. `server`
- `server_id` - (int) ID of the server which should be a target for this Load Balancer.

## Import

Load Balancers can be imported using its `id`:

```
terraform import hcloud_load_balancer.my_load_balancer <id>
```
