package placementgroup

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/placementgroup"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourcePlacementGroupTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := placementgroup.NewRData(t, "basic-placement-group", []int{4711}, "spread")
	res.SetRName("placement-group-ds-test")

	placementGroupByName := &placementgroup.DData{
		PlacementGroupName: res.TFID() + ".name",
	}
	placementGroupByName.SetRName("placement_group_by_name")

	placementGroupByID := &placementgroup.DData{
		PlacementGroupID: res.TFID() + ".id",
	}
	placementGroupByID.SetRName("placement_group_by_id")

	placementGroupBySel := &placementgroup.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	placementGroupBySel.SetRName("placement_group_by_sel")

	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(placementgroup.ResourceType, placementgroup.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", res,
					"testdata/d/hcloud_placement_group", placementGroupByName,
					"testdata/d/hcloud_placement_group", placementGroupByID,
					"testdata/d/hcloud_placement_group", placementGroupBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(placementGroupByName.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),

					resource.TestCheckResourceAttr(placementGroupByID.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),

					resource.TestCheckResourceAttr(placementGroupBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
				),
			},
		},
	})
}
