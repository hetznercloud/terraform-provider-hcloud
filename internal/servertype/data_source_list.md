Provides a list of available Hetzner Cloud Server Types.

## Example Usage

```hcl
data "hcloud_server_types" "all" {}

resource "hcloud_server" "workers" {
  count = 3

  name        = "node${count.index}"
  image       = "debian-12"
  server_type = element(data.hcloud_server_types.all.names, count.index)
}
```
