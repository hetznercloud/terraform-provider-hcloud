package hcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("data_source_floating_ip", &resource.Sweeper{
		Name: "hcloud_floating_ip_data_source",
		F:    testSweepFloatingIps,
	})
}

var floatingIPForDataSource *hcloud.FloatingIP

func TestAccHcloudDataSourceFloatingIP(t *testing.T) {
	floatingIPForDataSource, _ = createTestFloatingIP()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckFloatingIPDataSourceConfig(floatingIPForDataSource),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckFloatingIPExists("data.hcloud_floating_ip.ip_1", floatingIPForDataSource),
					resource.TestCheckResourceAttr(
						"data.hcloud_floating_ip.ip_1", "type", "ipv4"),
				),
			},
		},
	})

	testDataSourceCleaup()

}

func createTestFloatingIP() (*hcloud.FloatingIP, error) {
	client, _ := createClient()
	ctx := context.Background()
	description := "my-floating-ip.com"

	opts := hcloud.FloatingIPCreateOpts{
		Type:         hcloud.FloatingIPType(hcloud.FloatingIPTypeIPv4),
		Description:  hcloud.String(description),
		HomeLocation: &hcloud.Location{Name: "hel1"},
	}

	response, _, err := client.FloatingIP.Create(ctx, opts)
	if err != nil {
		return nil, err
	}

	return response.FloatingIP, nil
}

func testAccHcloudCheckFloatingIPDataSourceConfig(ip *hcloud.FloatingIP) string {
	return fmt.Sprintf(`
data "hcloud_floating_ip" "ip_1" {
  ip_address = "%s"
}`, ip.IP)
}

func testDataSourceCleaup() {
	testSweepFloatingIps("all")
}
