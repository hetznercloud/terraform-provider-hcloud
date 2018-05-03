---
layout: "hcloud"
page_title: "Provider: Hetzner Cloud"
sidebar_current: "docs-hcloud-index"
description: |-
  The Hetzner Cloud (hcloud) provider is used to interact with the resources supported by Hetzner Cloud.
---

# Hetzner Cloud Provider

The Hetzner Cloud (hcloud) provider is used to interact with the resources supported by Hetzner Cloud. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Set the variable value in *.tfvars file
# or using -var="hcloud_token=..." CLI option
variable "hcloud_token" {}

# Configure the Hetzner Cloud Provider
provider "hcloud" {
  token = "${var.hcloud_token}"
}

# Create a server
resource "hcloud_server" "web" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

- `token` - (Required) This is the Hetzner Cloud API Token. This can also be specified with the `HCLOUD_TOKEN` environment variable.
