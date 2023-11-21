Provides details about a specific Hetzner Cloud Location.

Use this resource to get detailed information about a specific location.

## Example

```hcl
data "hcloud_location" "location1" {
  name = "fsn1"
}

data "hcloud_location" "location2" {
  id = 1
}
```
