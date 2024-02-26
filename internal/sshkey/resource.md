Provides a Hetzner Cloud SSH key resource to manage SSH keys for server access.

## Example Usage

```hcl
# Create a new SSH key
resource "hcloud_ssh_key" "default" {
  name       = "Terraform Example"
  public_key = file("~/.ssh/id_rsa.pub")
}
```
