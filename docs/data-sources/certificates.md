---
page_title: "Hetzner Cloud: hcloud_certificates"
description: |-
  Provides details about multiple Hetzner Cloud Certificates.
---

# hcloud_certificates

Provides details about multiple Hetzner Cloud Certificates.

## Example Usage

```terraform
data "hcloud_certificates" "sample_certificate_1" {
  with_selector = "key=value"
}
```

## Argument Reference

- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/reference/cloud#label-selector)

## Attribute Reference

- `certificates` - (list) List of all matching certificates. See `data.hcloud_certificate` for schema.
