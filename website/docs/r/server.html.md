---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_server"
sidebar_current: "docs-hcloud-resource-server-x"
description: |-
  Provides an Hetzner Cloud server resource. This can be used to create, modify, and delete servers. Servers also support provisioning.
---

# hcloud_server

Provides an Hetzner Cloud server resource. This can be used to create, modify, and delete servers. Servers also support [provisioning](https://www.terraform.io/docs/provisioners/index.html).

## Example Usage

### Basic server creation

```hcl
# Create a new server running debian
resource "hcloud_server" "node1" {
  name        = "node1"
  image       = "debian-9"
  server_type = "cx11"
}
```

### Server creation with network

```hcl
resource "hcloud_network" "network" {
  name     = "network"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "network-subnet" {
  type         = "cloud"
  network_id   = hcloud_network.network.id
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

resource "hcloud_server" "server" {
  name        = "server"
  server_type = "cx11"
  image       = "ubuntu-20.04"
  location    = "nbg1"

  network {
    network_id = hcloud_network.network.id
    ip         = "10.0.1.5"
    alias_ips  = [
      "10.0.1.6",
      "10.0.1.7"
    ]
  }

  # **Note**: the depends_on is important when directly attaching the
  # server to a network. Otherwise Terraform will attempt to create
  # server and sub-network in parallel. This may result in the server
  # creation failing randomly.
  depends_on = [
    hcloud_network_subnet.network-subnet
  ]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required, string) Name of the server to create (must be unique per project and a valid hostname as per RFC 1123).
- `server_type` - (Required, string) Name of the server type this server should be created with.
- `image` - (Required, string) Name or ID of the image the server is created from.
- `location` - (Optional, string) The location name to create the server in. `nbg1`, `fsn1` or `hel1`
- `datacenter` - (Optional, string) The datacenter name to create the server in.
- `user_data` - (Optional, string) Cloud-Init user data to use during server creation
- `ssh_keys` - (Optional, list) SSH key IDs or names which should be injected into the server at creation time
- `keep_disk` - (Optional, bool) If true, do not upgrade the disk. This allows downgrading the server type later.
- `iso` - (Optional, string) ID or Name of an ISO image to mount.
- `rescue` - (Optional, string) Enable and boot in to the specified rescue system. This enables simple installation of custom operating systems. `linux64` `linux32` or `freebsd64`
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.
- `backups` - (Optional, boolean) Enable or disable backups.
- `firewall_ids` - (Optional, list) Firewall IDs the server should be attached to on creation.
- `network` - (Optional)  Network the server should be attached to on creation. (Can be specified multiple times)
- `placement_group_id` - (Optional, string) Placement Group ID the server added to on creation.
- `delete_protection` - (Optional, boolean) Enable or disable delete protection (Needs to be the same as `rebuild_protection`).
- `rebuild_protection` - (Optional, boolean) Enable or disable rebuild protection (Needs to be the same as `delete_protection`).

`network` support the following fields:
- `network_id` - (Required, int) ID of the network
- `ip` - (Optional, string) Specify the IP the server should get in the network
- `alias_ips` - (Optional, list) Alias IPs the server should have in the Network.


## Attributes Reference

The following attributes are exported:

- `id` - (int) Unique ID of the server.
- `name` - (string) Name of the server.
- `server_type` - (string) Name of the server type.
- `image` - (string) Name or ID of the image the server was created from.
- `location` - (string) The location name.
- `datacenter` - (string) The datacenter name.
- `backup_window` - (string) The backup window of the server, if enabled.
- `backups` - (boolean) Whether backups are enabled.
- `iso` - (string) ID or Name of the mounted ISO image.
- `ipv4_address` - (string) The IPv4 address.
- `ipv6_address` - (string) The first IPv6 address of the assigned network.
- `ipv6_network` - (string) The IPv6 network.
- `status` - (string) The status of the server.
- `labels` - (map) User-defined labels (key-value pairs)
- `network` - (map) Private Network the server shall be attached to.
  The Network that should be attached to the server requires at least
  one subnetwork. Subnetworks cannot be referenced by Servers in the
  Hetzner Cloud API. Therefore Terraform attempts to create the
  subnetwork in parallel to the server. This leads to a concurrency
  issue. It is therefore necessary to use `depends_on` to link the server
  to the respective subnetwork. See examples.
- `firewall_ids` - (Optional, list) Firewall IDs the server is attached to.
- `network` - (Optional, list)  Network the server should be attached to on creation. (Can be specified multiple times)
- `placement_group_id` - (Optional, string) Placement Group ID the server is assigned to.
- `delete_protection` - (boolean) Whether delete protection is enabled.
- `rebuild_protection` - (boolean) Whether rebuild protection is enabled.

a single entry in `network` support the following fields:
- `network_id` - (Required, int) ID of the network
- `ip` - (Optional, string) Specify the IP the server should get in the network
- `alias_ips` - (Optional, list) Alias IPs the server should have in the Network.
- `mac_address` - (Optional, string) The MAC address the private interface of the server has


## Import

Servers can be imported using the server `id`:

```
terraform import hcloud_server.myserver <id>
```
