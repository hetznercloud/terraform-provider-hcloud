package image_test

import (
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

const TestImageName = "ubuntu-20.04"
const TestImageId = "15512617"

func TestAccHcloudDataSourceImageTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	imageByName := &image.DData{
		ImageName: TestImageName,
	}
	imageByName.SetRName("image_by_name")
	imageByID := &image.DData{
		ImageID: TestImageId,
	}
	imageByID.SetRName("image_by_id")

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", imageByName,
					"testdata/d/hcloud_image", imageByID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(imageByName.TFID(),
						"name", TestImageName),
					resource.TestCheckResourceAttr(imageByName.TFID(), "id", TestImageId),

					resource.TestCheckResourceAttr(imageByID.TFID(),
						"name", TestImageName),
					resource.TestCheckResourceAttr(imageByID.TFID(), "id", TestImageId),
				),
			},
		},
	})
}
