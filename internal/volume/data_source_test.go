package volume_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"
)

func TestAccHcloudDataSourceVolumeTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &volume.RData{
		Name:         "some-volume",
		Size:         10,
		LocationName: "nbg1",
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	res.SetRName("some-volume")
	volByName := &volume.DData{
		Name:       "vol_by_name",
		VolumeName: res.TFID() + ".name",
	}
	volByID := &volume.DData{
		Name:     "vol_by_id",
		VolumeID: res.TFID() + ".id",
	}
	volBySel := &volume.DData{
		Name:          "vol_by_sel",
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(volume.ResourceType, volume.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_volume", res,
					"testdata/d/hcloud_volume", volByName,
					"testdata/d/hcloud_volume", volByID,
					"testdata/d/hcloud_volume", volBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(volByName.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(volByName.TFID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(volByName.TFID(), "size", strconv.Itoa(res.Size)),

					resource.TestCheckResourceAttr(volByID.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(volByID.TFID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(volByID.TFID(), "size", strconv.Itoa(res.Size)),

					resource.TestCheckResourceAttr(volBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(volBySel.TFID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(volBySel.TFID(), "size", strconv.Itoa(res.Size)),
				),
			},
		},
	})
}
