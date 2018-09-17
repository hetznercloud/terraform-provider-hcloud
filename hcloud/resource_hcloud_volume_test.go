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
	resource.AddTestSweepers("hcloud_floating_ip", &resource.Sweeper{
		Name: "hcloud_floating_ip",
		F:    testSweepFloatingIps,
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
						"hcloud_volume.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "size", "10"),
				),
			},
			{
				Config: testAccHcloudCheckVolumeConfig_resize(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume.foobar", &volume),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_volume.foobar", "size", "100"),
				),
			},
		},
	})
}

func testAccHcloudCheckVolumeConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_volume" "foobar" {
  name       = "foobar-%d"
  size       = 10
}
`, rInt)
}

func testAccHcloudCheckVolumeConfig_resize(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_volume" "foobar" {
  name       = "foobar-%d"
  size       = 100
}
`, rInt)
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
		if _, err := client.Volume.Delete(ctx, volume); err != nil {
			return err
		}
	}

	return nil
}
