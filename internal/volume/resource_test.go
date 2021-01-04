package volume_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/volume"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestVolumeResource_Basic(t *testing.T) {
	var vol hcloud.Volume

	res := volume.Basic
	resRenamed := &volume.RData{
		Name:         res.Name + "-renamed",
		LocationName: "nbg1",
		Size:         10,
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	resRenamed.SetRName(res.RName())

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(volume.ResourceType, volume.ByID(t, &vol)),
		Steps: []resource.TestStep{
			{
				// Create a new Volume using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_volume", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), volume.ByID(t, &vol)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-volume--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "size", "10"),
					resource.TestCheckResourceAttr(res.TFID(), "location", "nbg1"),
				),
			},
			{
				// Try to import the newly created volume
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the Volume created in the previous step by
				// setting all optional fields and renaming the volume.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_volume", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("basic-volume-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "size", "10"),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "location", "nbg1"),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "labels.key2", "value2"),
				),
			},
		},
	})
}

func TestVolumeResource_Resize(t *testing.T) {
	var vol hcloud.Volume

	res := volume.Basic
	res.Name = "resized-volume"
	resResized := &volume.RData{
		Name:         res.Name,
		LocationName: "nbg1",
		Size:         25,
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	resResized.SetRName(res.RName())

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(volume.ResourceType, volume.ByID(t, &vol)),
		Steps: []resource.TestStep{
			{
				// Create a new Volume using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_volume", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), volume.ByID(t, &vol)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("resized-volume--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "size", "10"),
					resource.TestCheckResourceAttr(res.TFID(), "location", "nbg1"),
				),
			},
			{
				// Update the Volume created in the previous step by
				// changing the size.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_volume", resResized,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resResized.TFID(), "name",
						fmt.Sprintf("resized-volume--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resResized.TFID(), "size", "25"),
					resource.TestCheckResourceAttr(resResized.TFID(), "location", "nbg1"),
					resource.TestCheckResourceAttr(resResized.TFID(), "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resResized.TFID(), "labels.key2", "value2"),
				),
			},
		},
	})
}

func TestVolumeResource_WithServer(t *testing.T) {
	var vol hcloud.Volume
	tmplMan := testtemplate.Manager{}
	resServer1 := &server.RData{
		Name:         "some-server",
		Type:         testsupport.TestServerType,
		Image:        testsupport.TestImage,
		LocationName: "nbg1",
	}
	resServer1.SetRName("some-server")

	resServer2 := &server.RData{
		Name:         "another-server",
		Type:         testsupport.TestServerType,
		Image:        testsupport.TestImage,
		LocationName: "nbg1",
	}
	resServer2.SetRName("another-server")

	res := volume.Basic
	res.Name = "volume-with-server"
	res.LocationName = ""
	res.ServerID = resServer1.TFID() + ".id"

	resAnotherServer := volume.Basic
	resAnotherServer.Name = "volume-with-server"
	resAnotherServer.LocationName = ""
	resAnotherServer.ServerID = resServer2.TFID() + ".id"
	resAnotherServer.SetRName(res.RName())

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(volume.ResourceType, volume.ByID(t, &vol)),
		Steps: []resource.TestStep{
			{
				// Create a new Volume using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer1,
					"testdata/r/hcloud_server", resServer2,
					"testdata/r/hcloud_volume", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), volume.ByID(t, &vol)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("volume-with-server--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "size", "10"),
					resource.TestCheckResourceAttr(res.TFID(), "location", "nbg1"),
				),
			},
			{
				// Try to import the newly created volume
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the Volume created in the previous step by
				// changing the attached server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer1,
					"testdata/r/hcloud_server", resServer2,
					"testdata/r/hcloud_volume", resAnotherServer),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), volume.ByID(t, &vol)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("volume-with-server--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "size", "10"),
					resource.TestCheckResourceAttr(res.TFID(), "location", "nbg1"),
				),
			},
		},
	})
}
