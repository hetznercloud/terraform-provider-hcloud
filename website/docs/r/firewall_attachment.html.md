---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_firewall_attachment"
sidebar_current: "docs-hcloud-resource-firewall-attachment"
description: |-
  Attaches resources to a Hetzner Cloud Firewall.
---

# hcloud_firewall_attachment

Attaches resource to a Hetzner Cloud Firewall.

*Note*: only one `hcloud_firewall_attachment` per Firewall is allowed.
Any resources that should be attached to that Firewall need to be
specified in that `hcloud_firewall_attachment`.

## Example Usage

### Attach Servers

```hcl
resource "hcloud_server" "test_server" {
    name        = "test-server"
    server_type = "cx11"
    image       = "ubuntu-20.04"
}

resource "hcloud_firewall" "basic_firewall" {
    name   = "basic_firewall"
}

resource "hcloud_firewall_attachment" "fw_ref" {
    firewall_id = hcloud_firewall.basic_firewall.id
    server_ids  = [hcloud_server.test_server.id]
}
```

### Attach Label Selectors

```hcl
resource "hcloud_server" "test_server" {
    name        = "test-server"
    server_type = "cx11"
    image       = "ubuntu-20.04"

    labels = {
      firewall-attachment = "test-server"
    }
}

resource "hcloud_firewall" "basic_firewall" {
    name = "basic_firewall"
}

resource "hcloud_firewall_attachment" "fw_ref" {
    firewall_id     = hcloud_firewall.basic_firewall.id
    label_selectors = ["firewall-attachment=test-server"]
}
```

### Ensure a server is attached to a Firewall on first boot

The `firewall_ids` property of the `hcloud_server` resource ensures that
a server is attached to the specified Firewalls before its first boot.
This is **not** the case when using the `hcloud_firewall_attachment`
resource to attach servers to a Firewall. In some scenarios this may
pose a security risk.

The following workaround ensures that a server is attached to a Firewall
*before* it first boots. However, the workaround requires two Firewalls.
Additionally the server resource definition needs to ignore any remote
changes to the `hcloud_server.firewall_ids` property. This is done using
the `ignore_remote_firewall_ids` property of `hcloud_server`.

```hcl
terraform {
  required_providers {
    hcloud = {
      source     = "hetznercloud/hcloud"
      version    = "1.32.2"
    }
  }
}

resource "hcloud_firewall" "deny_all" {
    name   = "deny_all"
}

resource "hcloud_server" "test_server" {
    name                       = "test-server"
    server_type                = "cx11"
    image                      = "ubuntu-20.04"
    ignore_remote_firewall_ids = true
    firewall_ids               = [
        hcloud_firewall.deny_all.id
    ]
}

resource "hcloud_firewall" "allow_rules" {
    name   = "allow_rules"

    rule {
        direction       = "in"
        protocol        = "tcp"
        port            = "22"
        source_ips      = [
            "0.0.0.0/0",
            "::/0",
        ]
        destination_ips = [
            format("%s/32", hcloud_server.test_server.ipv4_address)
        ]
    }
}

resource "hcloud_firewall_attachment" "deny_all_att" {
    firewall_id = hcloud_firewall.deny_all.id
    server_ids  = [hcloud_server.test_server.id]
}

resource "hcloud_firewall_attachment" "allow_rules_att" {
    firewall_id = hcloud_firewall.allow_rules.id
    server_ids  = [hcloud_server.test_server.id]
}
```

## Argument Reference

- `firewall_id` - (Required, int) ID of the firewall the resources
  should be attached to.
- `server_ids` - (Optional, List) List of Server IDs to attach to the
  firewall.
- `label_selectors` - (Optional, List) List of label selectors used to
  select resources to attach to the firewall.

## Attribute Reference

- `id` (int) - Unique ID representing this `hcloud_firewall_attachment`.
- `firewall_id` (int) - ID of the Firewall the resourced referenced by
  this attachment are attached to.
- `server_ids` (List) - List of Server IDs attached to the Firewall.
- `label_selectors` (List) - List of label selectors attached to the
  Firewall.
