package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const dataSSHKeysConfig = `
data "hcloud_ssh_keys" "keys_with_label" {
  with_selector = "foo=bar"
}
`

func TestAccHcloudDataSourceSSHKeys(t *testing.T) {
	rInt := acctest.RandInt()
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("hcloud-ds@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckSSHKeysDataSourceConfig(rInt, publicKeyMaterial),
			},
			{
				Config: dataSSHKeysConfig,
				Check: resource.TestCheckResourceAttr(
					"data.hcloud_ssh_keys.keys_with_label", "ssh_keys.0.name", fmt.Sprintf("sshkeys-%d", rInt)),
			},
		},
	})
}
func testAccHcloudCheckSSHKeysDataSourceConfig(rInt int, key string) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "sshkeys_ds" {
  name       = "sshkeys-%d"
  public_key = "%s"
  labels  = {
    foo = "bar"
  }
}
`, rInt, key)
}
