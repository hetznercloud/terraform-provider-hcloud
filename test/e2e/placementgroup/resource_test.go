package placementgroup

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/placementgroup"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestPlacementGroupResource(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		var g hcloud.PlacementGroup

		res := NewRData(t, "basic-placement-group", "spread")
		resRenamed := &RData{
			Name: res.Name + "-renamed",
			Type: "spread",
			Labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}
		resRenamed.SetRName(res.RName())

		updated := NewRData(t, "basic-placement-group", "spread")
		updated.SetRName(res.RName())
		tmplMan := testtemplate.Manager{}
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(placementgroup.ResourceType, ByID(t, &g)),
			Steps: []resource.TestStep{
				{
					// Create a new Placement Group using the required values
					// only.
					Config: tmplMan.Render(t, "testdata/r/hcloud_placement_group", res),
					Check: resource.ComposeTestCheckFunc(
						testsupport.CheckResourceExists(res.TFID(), ByID(t, &g)),
						resource.TestCheckResourceAttr(res.TFID(), "name",
							fmt.Sprintf("basic-placement-group--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(res.TFID(), "type", "spread"),
					),
				},
				{
					// Try to import the newly created Placement Group
					ResourceName:      res.TFID(),
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					// Update the Placement Group created in the previous step by
					// setting all optional fields and renaming the volume.
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_placement_group", resRenamed,
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
							fmt.Sprintf("basic-placement-group-renamed--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(resRenamed.TFID(), "labels.key1", "value1"),
						resource.TestCheckResourceAttr(resRenamed.TFID(), "labels.key2", "value2"),
					),
				},
			},
		})
	})
}
