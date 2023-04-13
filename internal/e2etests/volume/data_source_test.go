package volume_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"
)

func TestAccHcloudDataSourceVolumeTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &volume.RData{
		Name:         "some-volume",
		Size:         10,
		LocationName: e2etests.TestLocationName,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	res.SetRName("some-volume")

	volByName := &volume.DData{
		VolumeName: res.TFID() + ".name",
	}
	volByName.SetRName("vol_by_name")

	volByID := &volume.DData{
		VolumeID: res.TFID() + ".id",
	}
	volByID.SetRName("vol_by_id")

	volBySel := &volume.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	volBySel.SetRName("vol_by_sel")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(volume.ResourceType, volume.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_volume", res,
				),
			},
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

func TestAccHcloudDataSourceAttachedVolumeTest(t *testing.T) {
	var s hcloud.Server

	resServer := &server.RData{
		Name:  "volume-ds-test",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	resServer.SetRName("volume-ds-test")

	resVolume := &volume.RData{
		Name: "some-volume",
		Size: 10,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
		ServerID: resServer.TFID() + ".id",
	}
	resVolume.SetRName("some-volume")

	volByName := &volume.DData{
		VolumeName: resVolume.TFID() + ".name",
	}
	volByName.SetRName("vol_by_name")

	volByID := &volume.DData{
		VolumeID: resVolume.TFID() + ".id",
	}
	volByID.SetRName("vol_by_id")

	volBySel := &volume.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", resVolume.TFID()),
	}
	volBySel.SetRName("vol_by_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(volume.ResourceType, volume.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_volume", resVolume,
					"testdata/r/hcloud_server", resServer,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_volume", resVolume,
					"testdata/d/hcloud_volume", volByName,
					"testdata/d/hcloud_volume", volByID,
					"testdata/d/hcloud_volume", volBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &s)),

					resource.TestCheckResourceAttr(volByName.TFID(),
						"name", fmt.Sprintf("%s--%d", resVolume.Name, tmplMan.RandInt)),
					testsupport.CheckResourceAttrFunc(volBySel.TFID(), "server_id", func() string {
						return strconv.Itoa(s.ID)
					}),
					resource.TestCheckResourceAttr(volByName.TFID(), "size", strconv.Itoa(resVolume.Size)),

					resource.TestCheckResourceAttr(volByID.TFID(),
						"name", fmt.Sprintf("%s--%d", resVolume.Name, tmplMan.RandInt)),
					testsupport.CheckResourceAttrFunc(volBySel.TFID(), "server_id", func() string {
						return strconv.Itoa(s.ID)
					}),
					resource.TestCheckResourceAttr(volByID.TFID(), "size", strconv.Itoa(resVolume.Size)),

					resource.TestCheckResourceAttr(volBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", resVolume.Name, tmplMan.RandInt)),
					testsupport.CheckResourceAttrFunc(volBySel.TFID(), "server_id", func() string {
						return strconv.Itoa(s.ID)
					}),
					resource.TestCheckResourceAttr(volBySel.TFID(), "size", strconv.Itoa(resVolume.Size)),
				),
			},
		},
	})
}

func TestAccHcloudDataSourceVolumeListTest(t *testing.T) {
	res := &volume.RData{
		Name:         "volume-ds-test",
		Size:         10,
		LocationName: e2etests.TestLocationName,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	res.SetRName("volume-ds-test")

	volumesBySel := &volume.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	volumesBySel.SetRName("volumes_by_sel")

	allVolumesSel := &volume.DDataList{}
	allVolumesSel.SetRName("all_volumes_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(volume.ResourceType, volume.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_volume", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_volume", res,
					"testdata/d/hcloud_volumes", volumesBySel,
					"testdata/d/hcloud_volumes", allVolumesSel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(volumesBySel.TFID(), "volumes.*",
						map[string]string{
							"name":     fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"size":     strconv.Itoa(res.Size),
							"location": res.LocationName,
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(allVolumesSel.TFID(), "volumes.*",
						map[string]string{
							"name":     fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"size":     strconv.Itoa(res.Size),
							"location": res.LocationName,
						},
					),
				),
			},
		},
	})
}
