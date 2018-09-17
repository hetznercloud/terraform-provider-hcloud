---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_ssh_key"
sidebar_current: "docs-hcloud-datasource-ssh-key"
description: |-
  Provides details about a specific Hetzner Cloud SSH Key.
---
# Data Source: hcloud_sshkey
Provides details about a Hetzner Cloud SSH Key.
This resource is useful if you want to use a non-terraform managed SSH Key.
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
data "hcloud_ssh_key" "ssh_key_4" {
  selector = "key=value"
}
resource "hcloud_server" "main" {
  ssh_keys  = ["${data.hcloud_ssh_key.ssh_key_1.id}","${data.hcloud_ssh_key.ssh_key_2.id}","${data.hcloud_ssh_key.ssh_key_3.id}"]
}
```
## Argument Reference
- `id` - (Optional, string) ID of the SSH Key.
- `name` - (Optional, string) Name of the SSH Key.
- `fingerprint` - (Optional, string) Fingerprint of the SSH Key.
- `selector` - (Optional, string) Label selector for the [label selector](https://docs.hetzner.cloud/#overview-label-selector).

## Attributes Reference
- `id` - (int) Unique ID of the SSH Key.
- `name` - (string) Name of the SSH Key.
- `fingerprint` - (string) Fingerprint of the SSH Key.
- `public_key` - (string) Public Key of the SSH Key.
