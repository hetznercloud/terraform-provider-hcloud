---
subcategory: ""
page_title: "Use Hetzner Object Storage (S3)"
---

# Use Hetzner Object Storage (S3)

As of today, there's no native Hetzner Cloud API to manage the S3 buckets and credentials.
Therefore, the only supported method for managing the resources of a Hetzner Object Storage is via third-party providers.

We provide an [example workflow for creating buckets in the Hetzner documentation](https://docs.hetzner.com/storage/object-storage/getting-started/creating-a-bucket-minio-terraform/).

More information about the missing support can be found in [GitHub issue #1005](https://github.com/hetznercloud/terraform-provider-hcloud/issues/1005).
