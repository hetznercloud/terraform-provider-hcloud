package hcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAccHcloudDataSourceSSHKey(t *testing.T) {
	var key hcloud.SSHKey
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
				Config: testAccHcloudCheckSSHKeyDataSourceConfig(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckSSHKeyExists("hcloud_ssh_key.sshkey_ds", &key),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_1", "name", fmt.Sprintf("sshkey-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_1", "public_key", publicKeyMaterial),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_2", "name", fmt.Sprintf("sshkey-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_2", "public_key", publicKeyMaterial),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_3", "name", fmt.Sprintf("sshkey-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_3", "public_key", publicKeyMaterial),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_4", "name", fmt.Sprintf("sshkey-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.hcloud_ssh_key.ssh_4", "public_key", publicKeyMaterial),
				),
			},
		},
	})
}
func testAccHcloudCheckSSHKeyDataSourceConfig(rInt int, key string) string {
	return fmt.Sprintf(`
variable "labels" {
  type = "map"
  default = {
    "key" = "value"
  }
}
resource "hcloud_ssh_key" "sshkey_ds" {
  name       = "sshkey-%d"
  public_key = "%s"
  labels  = "${var.labels}"
}
data "hcloud_ssh_key" "ssh_1" {
  name = "${hcloud_ssh_key.sshkey_ds.name}"
}
data "hcloud_ssh_key" "ssh_2" {
  id =  "${hcloud_ssh_key.sshkey_ds.id}"
}
data "hcloud_ssh_key" "ssh_3" {
  fingerprint =  "${hcloud_ssh_key.sshkey_ds.fingerprint}"
}
data "hcloud_ssh_key" "ssh_4" {
  with_selector =  "key=${hcloud_ssh_key.sshkey_ds.labels["key"]}"
}
`, rInt, key)
}
