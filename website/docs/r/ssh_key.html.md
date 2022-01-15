---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_ssh_key"
sidebar_current: "docs-hcloud-resource-ssh-key"
description: |-
  Provides a Hetzner Cloud SSH key resource to manage SSH keys for server access.
---

# hcloud_ssh_key

Provides a Hetzner Cloud SSH key resource to manage SSH keys for server access.

## Example Usage

```hcl
# Create a new SSH key
resource "hcloud_ssh_key" "default" {
  name       = "Terraform Example"
  public_key = file("~/.ssh/id_rsa.pub")
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required, string) Name of the SSH key.
- `public_key` - (Required, string) The public key. If this is a file, it can be read using the file interpolation function
- `labels` - (Optional, map) User-defined labels (key-value pairs) should be created with.

## Attributes Reference

The following attributes are exported:

- `id` - (int) The unique ID of the key.
- `name` - (string) The name of the SSH key
- `public_key` - (string) The text of the public key
- `fingerprint` - (string) The fingerprint of the SSH key
- `labels` - (map) User-defined labels (key-value pairs)

## Import

SSH keys can be imported using the SSH key `id`:

```
terraform import hcloud_ssh_key.mykey <id>
```
