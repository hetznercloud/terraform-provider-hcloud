---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_primary_ip"
sidebar_current: "docs-hcloud-resource-primary-ip-x"
description: |-
  Provides a Hetzner Cloud Primary IP to represent a publicly-accessible static IP address that can be mapped to one of your servers.
---

# hcloud_primary_ip

Provides a Hetzner Cloud Primary IP to represent a publicly-accessible static IP address that can be mapped to one of your servers.

Currently, Primary IPs can be only attached to servers.

## Example Usage

```hcl
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
  server_type = "cx11"
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
- `id` - (int) Unique ID of the Primary IP.
- `type` - (string) Type of the Primary IP.
- `name` - (string) Name of the Primary IP.
- `auto_delete` - (boolean) Whether auto delete is enabled.
- `labels` - (string) Description of the Primary IP.
- `assignee_id` - (int) ID of the assigned resource
- `assignee_type` - (string, optional) The type of the assigned resource.
- `delete_protection` - (boolean) Whether delete protection is enabled.

## Attributes Reference
- `id` - (int) Unique ID of the Primary IP.
- `type` - (string) Type of the Primary IP.
- `name` - (string) Name of the Primary IP.
- `auto_delete` - (boolean) Whether auto delete is enabled.
- `labels` - (string) Description of the Primary IP.
- `ip_address` - (string) IP Address of the Primary IP.
- `assignee_id` - (int) ID of the assigned resource
- `assignee_type` - (string) The type of the assigned resource.
- `delete_protection` - (boolean) Whether delete protection is enabled.

## Import

Primary IPs can be imported using its `id`:

```
terraform import hcloud_primary_ip.myip <id>
```
