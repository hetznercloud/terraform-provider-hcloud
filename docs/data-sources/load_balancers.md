---
page_title: "Hetzner Cloud: hcloud_load_balancers"
description: |-
  Provides details about multiple Hetzner Cloud Load Balancers.
---

# hcloud_load_balancer

Provides details about multiple Hetzner Cloud Load Balancers.

## Example Usage

```terraform
data "hcloud_load_balancers" "lb_2" {

}
data "hcloud_load_balancers" "lb_3" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/reference/cloud#label-selector)

## Attributes Reference

- `load_balancers` - (list) List of all matching load balancers. See `data.hcloud_load_balancer` for schema.
