package image_test

import (
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceImageTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	imageByName := &image.DData{
		ImageName: teste2e.TestImage,
	}
	imageByName.SetRName("image_by_name")
	imageByID := &image.DData{
		ImageID: teste2e.TestImageID,
	}
	imageByID.SetRName("image_by_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", imageByName,
					"testdata/d/hcloud_image", imageByID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(imageByName.TFID(),
						"name", teste2e.TestImage),
					resource.TestCheckResourceAttr(imageByName.TFID(), "id", teste2e.TestImageID),

					resource.TestCheckResourceAttr(imageByID.TFID(),
						"name", teste2e.TestImage),
					resource.TestCheckResourceAttr(imageByID.TFID(), "id", teste2e.TestImageID),
				),
			},
		},
	})
}

func TestAccHcloudDataSourceImageWithFiltersTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	imageByName := &image.DData{
		ImageName:         teste2e.TestImage,
		Architecture:      "arm",
		IncludeDeprecated: true,
	}
	imageByName.SetRName("image_by_name")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_image", imageByName,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(imageByName.TFID(),
						"name", teste2e.TestImage),
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
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_images", allImagesSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(allImagesSel.TFID(), "images.*",
						map[string]string{
							"name": teste2e.TestImage,
							"id":   teste2e.TestImageID,
						},
					),
				),
			},
		},
	})
}
