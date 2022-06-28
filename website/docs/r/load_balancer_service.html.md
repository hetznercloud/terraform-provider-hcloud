---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_load_balancer_service"
sidebar_current: "docs-hcloud-resource-load-balancer-service-x"
description: |-
  Define services for Hetzner Cloud Load Balancers.
---

# hcloud_load_balancer_service

  Define services for Hetzner Cloud Load Balancers.

## Example Usage

```hcl
resource "hcloud_load_balancer" "load_balancer" {
  name               = "my-load-balancer"
  load_balancer_type = "lb11"
  location           = "nbg1"
}

resource "hcloud_load_balancer_service" "load_balancer_service" {
    load_balancer_id = hcloud_load_balancer.test_load_balancer.id
    protocol         = "http"
}
```

## Argument Reference

- `load_balancer_id` - (Required, string) Id of the load balancer this service belongs to.
- `protocol` - (Required, string) Protocol of the service. `http`, `https` or `tcp`
- `listen_port` - (Optional, int) Port the service listen on, required if protocol is `tcp`. Can be everything between `1` and `65535`. Must be unique per Load Balancer.
- `destination_port` - (Optional, int) Port the service connects to the targets on, required if protocol is `tcp`. Can be everything between `1` and `65535`.
- `proxyprotocol` - (Optional, bool) Enable proxyprotocol.
- `http` - (Optional, list) List of http configurations when `protocol` is `http` or `https`.
- `health_check` - (Optional, list) List of health check configurations when `protocol` is `http` or `https`.

`http` supports the following fields:

- `sticky_sessions` - (Optional, bool) Enable sticky sessions
- `cookie_name` - (Optional, string) Name of the cookie for sticky session. Default: `HCLBSTICKY`
- `cookie_lifetime` - (Optional, int) Lifetime of the cookie for sticky session (in seconds). Default: `300`
- `certificates` - (Optional, list[int]) List of IDs from certificates which the Load Balancer has.
- `redirect_http` - (Optional, bool) Redirect HTTP to HTTPS traffic. Only supported for services with `protocol` `https` using the default HTTP port `80`.

`health_check` supports the following fields:

- `protocol` - (Required, string) Protocol the health check uses. `http` or `tcp`
- `port` - (Required, int) Port the health check tries to connect to, required if protocol is `tcp`. Can be everything between `1` and `65535`. Must be unique per Load Balancer.
- `interval` - (Required, int) Interval how often the health check will be performed, in seconds.
- `timeout` - (Required, int) Timeout when a health check try will be canceled if there is no response, in seconds.
- `retries` - (Optional, int) Number of tries a health check will be performed until a target will be listed as `unhealthy`.
- `http` - (Optional, list) List of http configurations. Required if `protocol` is `http`.

(health check) `http` supports the following fields:

- `domain` - (Optional, string) Domain we try to access when performing the Health Check.
- `path` - (Optional, string) Path we try to access when performing the Health Check.
- `response` - (Optional, string) Response we expect to be included in the Target response when a Health Check was performed.
- `tls` - (Optional, bool) Enable TLS certificate checking.
- `status_codes` - (Optional, list[string]) We expect that the target answers with these status codes. If not the target is marked as `unhealthy`.


## Attribute Reference

- `protocol` - (string) Protocol of the service. `http`, `https` or `tcp`
- `listen_port` - (int) Port the service listen on. Can be everything between `1` and `65535`. Must be unique per Load Balancer.
- `destination_port` - (int) Port the service connects to the targets on. Can be everything between `1` and `65535`.
- `proxyprotocol` - (bool) Enable proxyprotocol.
- `http` - (list) List of http configurations when `protocol` is `http` or `https`.
- `health_check` - (list) List of http configurations when `protocol` is `http` or `https`.

`http` supports the following fields:

- `cookie_name` - (string) Name of the cookie for sticky session.
- `cookie_lifetime` - (int) Lifetime of the cookie for sticky session (in seconds).
- `certificates` - (list[int]) List of IDs from certificates which the Load Balancer has.

`health_check` supports the following fields:

- `protocol` - (string) Protocol the health check uses. `http`, `https` or `tcp`
- `port` - (int) Port the health check tries to connect to. Can be everything between `1` and `65535`.
- `interval` - (int) Interval how often the health check will be performed, in seconds.
- `timeout` - (int) Timeout when a health check try will be canceled if there is no response, in seconds.
- `retries` - (int) Number of tries a health check will be performed until a target will be listed as `unhealthy`.
- `http` - (list) List of http configurations when `protocol` is `http`.

(health check) `http` supports the following fields:

- `domain` -  string) Domain we try to access when performing the Health Check.
- `path` - (string) Path we try to access when performing the Health Check.
- `response` - (string) Response we expect to be included in the Target response when a Health Check was performed.
- `tls` - (bool) Enable TLS certificate checking.
- `status_codes` - (list[string]) We expect that the target answers with these status codes. If not the target is marked as `unhealthy`.

## Import

Load Balancer Service entries can be imported using a compound ID with the following format:
`<load-balancer-id>__<listen-port>`

```
terraform import hcloud_load_balancer_service.myloadbalancernetwork 123__80
```
