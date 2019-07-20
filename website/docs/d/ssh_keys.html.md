---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_ssh_keys"
sidebar_current: "docs-hcloud-datasource-ssh-keys"
description: |-
  Provides details about multiple Hetzner Cloud SSH Keys.
---

# Data Source: hcloud_sshkey

Provides details about Hetzner Cloud SSH Keys.
This resource is useful if you want to use a non-terraform managed SSH Key.

## Example Usage

```hcl
data "hcloud_ssh_keys" "all_keys" {
}
data "hcloud_ssh_keys" "keys_by_selector" {
  with_selector = "foo=bar"
}
resource "hcloud_server" "main" {
  ssh_keys  = "${data.hcloud_ssh_keys.all_keys.ssh_keys.*.name}"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/#overview-label-selector)

## Attributes Reference

- `ssh_keys` - (list) List of all matches SSH keys. See `data.hcloud_ssh_key` for schema.
