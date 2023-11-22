Provides a list of available Hetzner Cloud Datacenters.

This resource may be useful to create highly available infrastructure, distributed across several datacenters.

## Example

```hcl
data "hcloud_datacenters" "datacenters" {}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-11"
  server_type = "cx31"
  datacenter  = element(data.hcloud_datacenters.datacenters.datacenters, count.index).name
}
```
