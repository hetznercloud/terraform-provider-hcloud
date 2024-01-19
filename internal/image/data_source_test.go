package image_test

import (
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceImageTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	imageByName := &image.DData{
		ImageName: e2etests.TestImage,
	}
	imageByName.SetRName("image_by_name")
	imageByID := &image.DData{
		ImageID: e2etests.TestImageID,
	}
	imageByID.SetRName("image_by_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 e2etests.PreCheck(t),
		ProtoV6ProviderFactories: e2etests.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", imageByName,
					"testdata/d/hcloud_image", imageByID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(imageByName.TFID(),
						"name", e2etests.TestImage),
					resource.TestCheckResourceAttr(imageByName.TFID(), "id", e2etests.TestImageID),

					resource.TestCheckResourceAttr(imageByID.TFID(),
						"name", e2etests.TestImage),
					resource.TestCheckResourceAttr(imageByID.TFID(), "id", e2etests.TestImageID),
				),
			},
		},
	})
}

func TestAccHcloudDataSourceImageWithFiltersTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	imageByName := &image.DData{
		ImageName:         e2etests.TestImage,
		Architecture:      "arm",
		IncludeDeprecated: true,
	}
	imageByName.SetRName("image_by_name")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 e2etests.PreCheck(t),
		ProtoV6ProviderFactories: e2etests.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", imageByName,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(imageByName.TFID(),
						"name", e2etests.TestImage),
					resource.TestCheckResourceAttr(imageByName.TFID(),
						"architecture", "arm"),
				),
			},
		},
	})
}

func TestAccHcloudDataSourceImageListTest(t *testing.T) {
	allImagesSel := &image.DDataList{}
	allImagesSel.SetRName("all_images_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 e2etests.PreCheck(t),
		ProtoV6ProviderFactories: e2etests.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_images", allImagesSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(allImagesSel.TFID(), "images.*",
						map[string]string{
							"name": e2etests.TestImage,
							"id":   e2etests.TestImageID,
						},
					),
				),
			},
		},
	})
}
