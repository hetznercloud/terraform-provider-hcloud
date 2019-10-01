package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAccHcloudSSHKey_Basic(t *testing.T) {
	var key hcloud.SSHKey
	rInt := acctest.RandInt()
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("hcloud@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckSSHKeyConfig_basic(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckSSHKeyExists("hcloud_ssh_key.foobar", &key),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "name", fmt.Sprintf("foobar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "public_key", publicKeyMaterial),
				),
			},
		},
	})
}

func TestAccHcloudSSHKey_Update(t *testing.T) {
	var key hcloud.SSHKey
	rInt := acctest.RandInt()
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("hcloud@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckSSHKeyConfig_basic(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckSSHKeyExists("hcloud_ssh_key.foobar", &key),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "name", fmt.Sprintf("foobar-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "public_key", publicKeyMaterial),
				),
			},
			{
				Config: testAccHcloudCheckSSHKeyConfig_update(rInt, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckSSHKeyExists("hcloud_ssh_key.foobar", &key),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "name", fmt.Sprintf("foobar-updated-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_ssh_key.foobar", "public_key", publicKeyMaterial),
				),
			},
		},
	})
}

func testAccHcloudCheckSSHKeyDestroy(s *terraform.State) error {
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
			return fmt.Errorf("Error checking if SSH Key (%s) is deleted: %v", rs.Primary.ID, err)
		}
		if sshKey != nil {
			return fmt.Errorf("SSH key (%s) has not been deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccHcloudCheckSSHKeyExists(n string, key *hcloud.SSHKey) resource.TestCheckFunc {
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

func testAccHcloudCheckSSHKeyConfig_basic(rInt int, key string) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
    name = "foobar-%d"
    public_key = "%s"
}`, rInt, key)
}

func testAccHcloudCheckSSHKeyConfig_update(rInt int, key string) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
		name = "foobar-updated-%d"
		public_key = "%s"
}`, rInt, key)
}
