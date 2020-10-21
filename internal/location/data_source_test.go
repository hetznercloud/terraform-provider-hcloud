package location_test

import (
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/location"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
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

func TestAccHcloudDataSourceLocationsTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	locationsDS := &location.LocationsDData{}
	locationsDS.SetRName("ds")
	resource.Test(t, resource.TestCase{
		PreCheck:  testsupport.AccTestPreCheck(t),
		Providers: testsupport.AccTestProviders(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_locations", locationsDS,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(locationsDS.TFID(), "location_ids.0", "1"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "location_ids.1", "2"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "location_ids.2", "3"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "names.0", "fsn1"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "names.1", "nbg1"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "names.2", "hel1"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "descriptions.0", "Falkenstein DC Park 1"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "descriptions.1", "Nuremberg DC Park 1"),
					resource.TestCheckResourceAttr(locationsDS.TFID(), "descriptions.2", "Helsinki DC Park 1"),
				),
			},
		},
	})
}
