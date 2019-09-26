package hcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccHcloudFloatingIP_importServer(t *testing.T) {
	resourceName := "hcloud_floating_ip.floating_ip"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPConfig_server(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
