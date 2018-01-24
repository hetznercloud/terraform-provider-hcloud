# hcloud_ssh_key

## Example Usage

```
# Create a new SSH key
resource "hcloud_ssh_key" "default" {
  name = "Terraform Example"
  public_key = "${file("~/.ssh/id_rsa.pub")}"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the SSH key.
- `public_key` - (Required) The public key. If this is a file, it can be read using the file interpolation function

## Attributes Reference

The following attributes are exported:

- `id` - The unique ID of the key.
- `name` - The name of the SSH key
- `public_key` - The text of the public key
- `fingerprint` - The fingerprint of the SSH key
