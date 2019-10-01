package hcloud

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("data_source_image", &resource.Sweeper{
		Name: "hcloud_image_data_source",
	})
}
func TestAccHcloudDataSourceImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccHcloudPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckImageDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_1", "type", "system"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_1", "os_flavor", "ubuntu"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_1", "description", "Ubuntu 18.04"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_2", "type", "system"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_2", "os_flavor", "ubuntu"),
					resource.TestCheckResourceAttr(
						"data.hcloud_image.image_2", "description", "Ubuntu 18.04"),
				),
			},
		},
	})
}
func TestAccHcloudDataSourceImageSort(t *testing.T) {
	testImageList := []*hcloud.Image{
		{
			ID:      5,
			Created: time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:      6,
			Created: time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:      7,
			Created: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		}}
	oldestImage := testImageList[2]
	sortImageListByCreated(testImageList)
	if testImageList[0].ID != oldestImage.ID {
		t.Fatalf("sortImageListByCreated did not sort by date: expected %d/%s but got %d/%s", oldestImage.ID, oldestImage.Created, testImageList[0].ID, testImageList[0].Created)
	}
}

func testAccHcloudCheckImageDataSourceConfig() string {
	return fmt.Sprintf(`
data "hcloud_image" "image_1" {
  name = "ubuntu-18.04"
}
data "hcloud_image" "image_2" {
  id = 168855
}
`)
}
