---
page_title: "Hetzner Cloud: hcloud_load_certificate"
description: |-
  Provides details about a specific Hetzner Cloud Certificate.
---

# hcloud_certificate

Provides details about a specific Hetzner Cloud Certificate.

```terraform
data "hcloud_certificate" "sample_certificate_1" {
  name = "sample-certificate-1"
}

data "hcloud_certificate" "sample_certificate_2" {
  id = "4711"
}
```

## Argument Reference

- `id` - ID of the certificate.
- `name` - Name of the certificate.
- `with_selector` - (Optional, string) [Label selector](https://docs.hetzner.cloud/reference/cloud#label-selector)

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
