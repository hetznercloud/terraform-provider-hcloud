Provides details about a specific Hetzner Cloud Server Type.

Use this resource to get detailed information about specific Server Type.

## Example Usage

```hcl
data "hcloud_server_type" "by_name" {
  name = "cx22"
}

data "hcloud_server_type" "by_id" {
  id = 1
}
```
