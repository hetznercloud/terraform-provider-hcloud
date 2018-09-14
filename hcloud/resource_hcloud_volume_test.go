package hcloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

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
