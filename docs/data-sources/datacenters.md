---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hcloud_datacenters Data Source - hcloud"
subcategory: ""
description: |-
  Provides a list of available Hetzner Cloud Datacenters.
  This resource may be useful to create highly available infrastructure, distributed across several Datacenters.
---

# hcloud_datacenters (Data Source)

Provides a list of available Hetzner Cloud Datacenters.

This resource may be useful to create highly available infrastructure, distributed across several Datacenters.

## Example Usage

```terraform
data "hcloud_datacenters" "all" {}

resource "hcloud_server" "workers" {
  count = 5

  name        = "node${count.index}"
  image       = "debian-12"
  server_type = "cx22"
  datacenter  = element(data.hcloud_datacenters.all.datacenters, count.index).name
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `datacenter_ids` (List of String, Deprecated)
- `datacenters` (Attributes List) (see [below for nested schema](#nestedatt--datacenters))
- `descriptions` (List of String, Deprecated)
- `id` (String) The ID of this resource.
- `names` (List of String, Deprecated)

<a id="nestedatt--datacenters"></a>
### Nested Schema for `datacenters`

Read-Only:

- `available_server_type_ids` (List of Number) List of currently available Server Types in the Datacenter.
- `description` (String) Description of the Datacenter.
- `id` (Number) ID of the Datacenter.
- `location` (Map of String) Location of the Datacenter. See the [Hetzner Docs](https://docs.hetzner.com/cloud/general/locations/#what-locations-are-there) for more details about locations.
- `name` (String) Name of the Datacenter.
- `supported_server_type_ids` (List of Number) List of supported Server Types in the Datacenter.