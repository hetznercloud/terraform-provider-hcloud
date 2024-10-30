Provides details about a specific Hetzner Cloud Load Balancer Type.

Use this resource to get detailed information about specific Load Balancer Type.

## Example Usage

```hcl
data "hcloud_load_balancer_type" "by_name" {
  name = "lb11"
}

data "hcloud_load_balancer_type" "by_id" {
  id = 1
}
```
