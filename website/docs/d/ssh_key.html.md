---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_ssh_key"
sidebar_current: "docs-hcloud-datasource-ssh-key"
description: |-
  Provides details about a specific Hetzner Cloud SSH Key.
---
# Data Source: hcloud_sshkey
Provides details about a Hetzner Cloud SSH Key.
This resource can be useful when you need to determine a SSH Key, that is not managed by terraform.
## Example Usage
```hcl
data "hcloud_ssh_key" "ssh_key_1" {
  id = "1234"
}
data "hcloud_ssh_key" "ssh_key_2" {
  name = "my-ssh-key"
}
data "hcloud_ssh_key" "ssh_key_3" {
  fingerprint = "43:51:43:a1:b5:fc:8b:b7:0a:3a:a9:b1:0f:66:73:a8"
}
resource "hcloud_server" "main" {
  ssh_keys  = ["${data.hcloud_ssh_key.ssh_key_1.id}","${data.hcloud_ssh_key.ssh_key_2.id}","${data.hcloud_ssh_key.ssh_key_3.id}"]
}
```
## Argument Reference
- `id` - ID of the SSH Key.
- `name` - Name of the SSH Key.
- `fingerprint` - Fingerprint of the SSH Key.
## Attributes Reference
- `id` - Unique ID of the SSH Key.
- `name` - Name of the SSH Key.
- `fingerprint` - Fingerprint of the SSH Key.
- `public_key` - Public Key of the SSH Key.