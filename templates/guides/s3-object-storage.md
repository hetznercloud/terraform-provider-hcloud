---
subcategory: ""
page_title: "Use Hetzner Object Storage (S3)"
---

# Use Hetzner Object Storage (S3) with the `aminueza/minio` provider

As of today, there's no native Hetzner Cloud API to manage the S3 buckets and credentials.
Therefore, the only supported method for managing the resources of a Hetzner Object Storage is via third-party providers.

We provide an [example workflow for creating new buckets in the Hetzner documentation](https://docs.hetzner.com/storage/object-storage/getting-started/creating-a-bucket-minio-terraform/).

## Note about importing resources

To import existing S3 buckets to your state, you must use the name of the bucket as the `id`.
Example `import` block:

```hcl
import {
  to = minio_s3_bucket.my_bucket
  id = "my-bucket-h7892dd9"
}
```
