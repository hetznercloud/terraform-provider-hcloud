---
layout: "hcloud"
page_title: "Hetzner Cloud: hcloud_managed_certificate"
sidebar_current: "docs-hcloud-resource-managed-certificate-x"
description: |-
  Obtain a TLS Certificate managed by Hetzner Cloud.
---

# hcloud_managed_certificate

Obtain a Hetzner Cloud managed TLS certificate.

## Example Usage

```hcl
resource "hcloud_managed_certificate" "managed_cert" {
  name         = "managed_cert"
  domain_names = ["*.example.com", "example.com"]
  labels = {
    label_1 = "value_1"
    label_2 = "value_2"
    ...
  }
}
```

## Import

Managed certificates can be imported using their `id`:

```hcl
terraform import hcloud_managed_certificate.sample_certificate id
```

## Argument Reference

- `name` - (Required, string) Name of the Certificate.
- `domains` - (Required, list) Domain names for which a certificate
  should be obtained.
- `labels` - (Optional, map) User-defined labels (key-value pairs) the
certificate should be created with.

## Attribute Reference

- `id` - (int) Unique ID of the certificate.
- `name` - (string) Name of the Certificate.
- `certificate` - (string) PEM encoded TLS certificate.
- `labels` - (map) User-defined labels (key-value pairs) assigned to the certificate.
- `domain_names` - (list) Domains and subdomains covered by the certificate.
- `fingerprint` - (string) Fingerprint of the certificate.
- `created` - (string) Point in time when the Certificate was created at Hetzner Cloud (in ISO-8601 format).
- `not_valid_before` - (string) Point in time when the Certificate becomes valid (in ISO-8601 format).
- `not_valid_after` - (string) Point in time when the Certificate stops being valid (in ISO-8601 format).
