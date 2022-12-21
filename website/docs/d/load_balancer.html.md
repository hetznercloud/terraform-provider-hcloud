---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_load_balancer"
sidebar_current: "docs-hcloud-datasource-load-balancer-x"
description: |-
   Provides details about a specific Hetzner Cloud Load Balancer.
---

# hcloud_load_balancer

    Provides details about a specific Hetzner Cloud Load Balancer.

## Example Usage

```hcl
data "hcloud_load_balancer" "lb_1" {
  name = "my-load-balancer"
}
data "hcloud_load_balancer" "lb_2" {
  id = "123"
}
data "hcloud_load_balancer" "lb_3" {
  with_selector = "key=value"
}
```

## Argument Reference
- `id` - ID of the Load Balancer.
- `name` - Name of the Load Balancer.
- `with_selector` - Label Selector. For more information about possible values, visit the [Hetzner Cloud Documentation](https://docs.hetzner.cloud/#overview-label-selector).

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
- `delete_protection` - (bool) Whether delete protection is enabled.
- `network_id` - (int) ID of the first private network that this Load Balancer is connected to.
- `network_ip` - (string) IP of the Load Balancer in the first private network that it is connected to.

`algorithm` support the following fields:
- `type` - (string) Type of the Load Balancer Algorithm. `round_robin` or `least_connection`

`target` support the following fields:
- `type` - (string) Type of the target. `server` or `label_selector`
- `server_id` - (int) ID of the server which should be a target for this Load Balancer.
- `label_selector` - (string) Label Selector to add a group of resources based on the label.

`service` support the following fields:
- `protocol` - (string) Protocol of the service. `http`, `https` or `tcp`
- `listen_port` - (int) Port the service listen on`. Can be everything between `1` and `65535`. Must be unique per Load Balancer.
- `destination_port` - (int) Port the service connects to the targets on. Can be everything between `1` and `65535`.
- `proxyprotocol` - (bool) Enable proxyprotocol.
- `http` - (list) List of http configurations when `protocol` is `http` or `https`.
- `health_check` - (list) List of http configurations when `protocol` is `http` or `https`.

`http` support the following fields:
- `sticky_sessions` - (string) Determine if sticky sessions are enabled or not.
- `cookie_name` - (string) Name of the cookie for sticky session.
- `cookie_lifetime` - (int) Lifetime of the cookie for sticky session (in seconds).
- `certificates` - (list[int]) List of IDs from certificates which the Load Balancer has.
- `redirect_http` - (string) Determine if all requests from port 80 should be redirected to port 443.

`health_check` support the following fields:
- `protocol` - (string) Protocol the health check uses. `http`, `https` or `tcp`
- `port` - (int) Port the health check tries to connect to. Can be everything between `1` and `65535`.
- `interval` - (int) Interval how often the health check will be performed, in seconds.
- `timeout` - (int) Timeout when a health check try will be canceled if there is no response, in seconds.
- `retries` - (int) Number of tries a health check will be performed until a target will be listed as `unhealthy`.
- `http` - (list) List of http configurations when `protocol` is `http` or `https`.

(health check) `http` support the following fields:
- `domain` -  string) Domain we try to access when performing the Health Check.
- `path` - (string) Path we try to access when performing the Health Check.
- `response` - (string) Response we expect to be included in the Target response when a Health Check was performed.
- `tls` - (bool) Enable TLS certificate checking.
- `status_codes` - (list[int]) We expect that the target answers with these status codes. If not the target is marked as `unhealthy`.


## Import

Load Balancers can be imported using its `id`:

```
terraform import hcloud_load_balancer.my_load_balancer id
```
