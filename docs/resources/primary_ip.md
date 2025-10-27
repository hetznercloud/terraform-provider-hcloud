---
page_title: "Hetzner Cloud: hcloud_primary_ip"
description: |-
  Provides a Hetzner Cloud Primary IP to represent a publicly-accessible static IP address that can be mapped to one of your servers.
---

# hcloud_primary_ip

Provides a Hetzner Cloud Primary IP to represent a publicly-accessible static IP address that can be mapped to one of your servers.

If a server is getting created, it has to have a primary ip. If a server is getting created without defining primary ips, two of them (one ipv4 and one ipv6) getting created & attached.
Currently, Primary IPs can be only attached to servers.

## Example Usage

```terraform
resource "hcloud_primary_ip" "main" {
  name          = "primary_ip_test"
  datacenter    = "fsn1-dc14"
  type          = "ipv4"
  assignee_type = "server"
  auto_delete   = true
  labels = {
    "hallo" : "welt"
  }
}
// Link a server to a primary IP
resource "hcloud_server" "server_test" {
  name        = "test-server"
  image       = "ubuntu-20.04"
  server_type = "cx23"
  datacenter  = "fsn1-dc14"
  labels = {
    "test" : "tessst1"
  }
  public_net {
    ipv4 = hcloud_primary_ip.main.id
  }

}
```

## Argument Reference

- `type` - (string) Type of the Primary IP. `ipv4` or `ipv6`
- `name` - (string) Name of the Primary IP.
-
`datacenter` - (string, optional) The datacenter name to create the resource in. See the [Hetzner Docs](https://docs.hetzner.com/cloud/general/locations/#what-datacenters-are-there) for more details about datacenters.
- `auto_delete` - (bool) Whether auto delete is enabled.
  `Important note:`It is recommended to set `auto_delete` to
  `false`, because if a server assigned to the managed ip is getting deleted, it will also delete the primary IP which will break the TF state.
- `labels` - (map, optional) User-defined labels (key-value pairs).
- `assignee_id` - (int, optional) ID of the assigned resource.
-
`assignee_type` - (string) The type of the assigned resource. Currently supported:
`server`
-
`delete_protection` - (bool, optional) Whether delete protection is enabled. See ["Delete Protection"](../index.html.markdown#delete-protection) in the Provider Docs for details.

Note: At least one of `datacenter` or `assignee_id` is required.

## Attributes Reference

- `id` - (int) Unique ID of the Primary IP.
- `type` - (string) Type of the Primary IP.
-
`datacenter` - (string) The datacenter of the Primary IP. See the [Hetzner Docs](https://docs.hetzner.com/cloud/general/locations/#what-datacenters-are-there) for more details about datacenters.
- `name` - (string) Name of the Primary IP.
- `auto_delete` - (bool) Whether auto delete is enabled.
- `labels` - (map) User-defined labels (key-value pairs).
- `ip_address` - (string) IP Address of the Primary IP.
-
`ip_network` - (string) IPv6 subnet of the Primary IP for IPv6 addresses. (Only set if
`type` is `ipv6`)
- `assignee_id` - (int) ID of the assigned resource.
- `assignee_type` - (string) The type of the assigned resource.
- `delete_protection` - (bool) Whether delete protection is enabled.

## Import

Primary IPs can be imported using its `id`:

```shell
terraform import hcloud_primary_ip.example "$PRIMARY_IP_ID"
```
