package datacenter_test

import (
	"github.com/terraform-providers/terraform-provider-hcloud/internal/datacenter"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceDatacenterTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	dcByName := &datacenter.DData{
		DatacenterName: "fsn1-dc14",
	}
	dcByName.SetRName("dc_by_name")
	dcByID := &datacenter.DData{
		DatacenterID: "4",
	}
	dcByID.SetRName("dc_by_id")
	resource.Test(t, resource.TestCase{
		PreCheck:  testsupport.AccTestPreCheck(t),
		Providers: testsupport.AccTestProviders(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_datacenter", dcByName,
					"testdata/d/hcloud_datacenter", dcByID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dcByName.TFID(), "id", "4"),
					resource.TestCheckResourceAttr(dcByName.TFID(), "name", "fsn1-dc14"),
					resource.TestCheckResourceAttr(dcByName.TFID(), "description", "Falkenstein 1 DC14"),

					resource.TestCheckResourceAttr(dcByID.TFID(), "id", "4"),
					resource.TestCheckResourceAttr(dcByID.TFID(), "name", "fsn1-dc14"),
					resource.TestCheckResourceAttr(dcByID.TFID(), "description", "Falkenstein 1 DC14"),
				),
			},
		},
	})
}
