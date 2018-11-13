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

func init() {
	resource.AddTestSweepers("hcloud_volume", &resource.Sweeper{
		Name: "hcloud_volume",
		F:    testSweepVolumes,
	})
}

func TestAccHcloudVolume_Basic(t *testing.T) {
	var volume hcloud.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckVolumeConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume.foobar", &volume),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "name", fmt.Sprintf("foo-volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "size", "10"),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "location", "nbg1"),
				),
			},
			{
				Config: testAccHcloudCheckVolumeConfig_resize(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume.foobar", &volume),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "name", fmt.Sprintf("foo-volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "size", "15"),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "location", "nbg1"),
				),
			},
			{
				Config: testAccHcloudCheckVolumeConfig_WithServer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume.foobar", &volume),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "name", fmt.Sprintf("foo-volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "size", "15"),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "location", "nbg1"),
				),
			},
			{
				Config: testAccHcloudCheckVolumeConfig_WithAnotherServer(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume.foobar", &volume),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "name", fmt.Sprintf("foo-volume-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "size", "15"),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "location", "nbg1"),
				),
			},
		},
	})
}

func testAccHcloudCheckVolumeConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_volume" "foobar" {
  name       = "foo-volume-%d"
  size       = 10
  location   = "nbg1"
}
`, rInt)
}

func testAccHcloudCheckVolumeConfig_resize(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_volume" "foobar" {
  name       = "foo-volume-%d"
  size       = 15
  location   = "nbg1"
}
`, rInt)
}

func testAccHcloudCheckVolumeConfig_WithServer(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "server_volume_foobar" {
  name        = "foo-volume-server-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "nbg1-dc3"
}
resource "hcloud_volume" "foobar" {
  name       = "foo-volume-%d"
  size       = 15
  server_id  = "${hcloud_server.server_volume_foobar.id}"
}
`, rInt, rInt)
}

func testAccHcloudCheckVolumeConfig_WithAnotherServer(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_server" "server_another_volume_foobar" {
  name        = "foo-volume-server-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "nbg1-dc3"
}
resource "hcloud_volume" "foobar" {
  name       = "foo-volume-%d"
  size       = 15
  server_id  = "${hcloud_server.server_another_volume_foobar.id}"
}
`, rInt, rInt)
}

func testAccHcloudCheckVolumeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_volume" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("volume id is no int: %v", err)
		}
		var volume *hcloud.Volume
		volume, _, err = client.Volume.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error checking if volume (%s) is deleted: %v",
				rs.Primary.ID, err)
		}
		if volume != nil {
			return fmt.Errorf("volume (%s) has not been deleted", rs.Primary.ID)
		}
	}

	return nil
}
func testAccHcloudCheckVolumeExists(n string, volume *hcloud.Volume) resource.TestCheckFunc {
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
		foundVolume, _, err := client.Volume.GetByID(context.Background(), id)
		if err != nil {
			return err
		}

		if foundVolume == nil {
			return fmt.Errorf("Record not found")
		}

		*volume = *foundVolume
		return nil
	}
}

func testSweepVolumes(region string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	volumes, err := client.Volume.All(ctx)
	if err != nil {
		return err
	}

	for _, volume := range volumes {
		if volume.Server != nil {
			action, _, _ := client.Volume.Detach(ctx, volume)
			if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
				return err
			}
		}
		if _, err := client.Volume.Delete(ctx, volume); err != nil {
			return err
		}
	}

	return nil
}
