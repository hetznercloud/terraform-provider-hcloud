package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAccSSHKey_Basic(t *testing.T) {
	var key hcloud.SSHKey
	rInt := acctest.RandInt()
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("hcloud@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSSHKeyConfig_basic(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSHKeyExists("hcloud_ssh_key.foobar", &key),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "name", fmt.Sprintf("foobar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "public_key", publicKeyMaterial),
				),
			},
		},
	})
}

func testAccCheckSSHKeyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_ssh_key" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		sshKey, _, err := client.SSHKey.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf("Error checking if SSH Key is deleted: %v", err)
		}
		if sshKey != nil {
			return fmt.Errorf("SSH key still exists")
		}
	}

	return nil
}

func testAccCheckSSHKeyExists(n string, key *hcloud.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Try to find the key
		foundKey, _, err := client.SSHKey.GetByID(context.Background(), id)
		if err != nil {
			return err
		}

		if foundKey == nil {
			return fmt.Errorf("Record not found")
		}

		*key = *foundKey
		return nil
	}
}

func testAccCheckSSHKeyConfig_basic(rInt int, key string) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
    name = "foobar-%d"
    public_key = "%s"
}`, rInt, key)
}
