package location_test

import (
	"github.com/terraform-providers/terraform-provider-hcloud/internal/location"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceLocationTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	lByName := &location.DData{
		LocationName: "fsn1",
	}
	lByName.SetRName("l_by_name")
	lByID := &location.DData{
		LocationID: "1",
	}
	lByID.SetRName("l_by_id")
	resource.Test(t, resource.TestCase{
		PreCheck:  testsupport.AccTestPreCheck(t),
		Providers: testsupport.AccTestProviders(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_location", lByName,
					"testdata/d/hcloud_location", lByID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lByName.TFID(), "id", "1"),
					resource.TestCheckResourceAttr(lByName.TFID(), "name", "fsn1"),
					resource.TestCheckResourceAttr(lByName.TFID(), "description", "Falkenstein DC Park 1"),

					resource.TestCheckResourceAttr(lByID.TFID(), "id", "1"),
					resource.TestCheckResourceAttr(lByID.TFID(), "name", "fsn1"),
					resource.TestCheckResourceAttr(lByID.TFID(), "description", "Falkenstein DC Park 1"),
				),
			},
		},
	})
}
