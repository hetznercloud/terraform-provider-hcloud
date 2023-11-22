Provides details about a specific Hetzner Cloud Datacenter. 

Use this resource to get detailed information about a specific datacenter.

## Example

```hcl
data "hcloud_datacenter" "datacenter1" {
  name = "fsn1-dc14"
}

data "hcloud_datacenter" "datacenter2" {
  id = 4
}
```
