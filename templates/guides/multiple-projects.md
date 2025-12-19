---
subcategory: ""
page_title: "Manage resources across multiple projects"
---

# Manage resources across multiple projects

In some scenarios, it is useful to manage you Hetzner Cloud resources across multiple projects.

Below is an example that demonstrate how to manage a DNS Zone in one project, while configuring the DNS Zone using resources from another project:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.57.0"
    }
  }
}

// Provider to manage the dns project
provider "hcloud" {
  // Defines the "alias" for the provider attached to the dns project.
  alias = "dns"
  token = "<token for the dns project>"
}

// Provider to manage another project
provider "hcloud" {
  // Defines the "alias" for the provider attached to another project.
  alias = "another"
  token = "<token for the another project>"
}

resource "hcloud_server" "host1" {
  // This server in managed in "another" project.
  provider = hcloud.another

  name        = "host1"
  location    = "hel1"
  server_type = "cpx22"
  image       = "debian-13"
}


resource "hcloud_zone" "main" {
  // The zone is managed in the "dns" project.
  provider = hcloud.dns

  name = "example.com"
  mode = "primary"
}

resource "hcloud_zone_rrset" "host1_a" {
  provider = hcloud.dns

  zone = hcloud_zone.main.name
  name = "host1"
  type = "A"
  records = [
    // The record is managed in the "dns" project, but the record values
    // are taken from a resource in another project.
    { value = hcloud_server.host1.ipv4_address }
  ]
}
```
