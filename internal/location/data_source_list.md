Provides a list of available Hetzner Cloud Locations.

This resource may be useful to create highly available infrastructure, distributed across several locations.

## Example

```hcl
data "hcloud_locations" "locations" {}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-11"
  server_type = "cx22"
  location    = element(data.hcloud_locations.locations.locations, count.index).name
}
```
