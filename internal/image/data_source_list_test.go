package image_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccImageDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	all := &image.DDataList{}
	all.SetRName("all")

	byArchitecture := &image.DDataList{
		Architecture: "arm",
	}
	byArchitecture.SetRName("architecture")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_images", all,
					"testdata/d/hcloud_images", byArchitecture,
				),

				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(all.TFID(), tfjsonpath.New("images"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(all.TFID(), tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("name"), knownvalue.NotNull()),

					statecheck.ExpectKnownValue(byArchitecture.TFID(), tfjsonpath.New("images"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(byArchitecture.TFID(), tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("name"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(byArchitecture.TFID(), tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("architecture"), knownvalue.StringExact("arm")),
				},
			},
		},
	})
}
