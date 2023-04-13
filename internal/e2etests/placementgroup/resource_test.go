package placementgroup

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/placementgroup"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestPlacementGroupResource_Basic(t *testing.T) {
	var g hcloud.PlacementGroup

	res := placementgroup.NewRData(t, "basic-placement-group", "spread")
	resRenamed := &placementgroup.RData{
		Name: res.Name + "-renamed",
		Type: "spread",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	resRenamed.SetRName(res.RName())

	updated := placementgroup.NewRData(t, "basic-placement-group", "spread")
	updated.SetRName(res.RName())
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(placementgroup.ResourceType, placementgroup.ByID(t, &g)),
		Steps: []resource.TestStep{
			{
				// Create a new Placement Group using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_placement_group", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), placementgroup.ByID(t, &g)),
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
}
