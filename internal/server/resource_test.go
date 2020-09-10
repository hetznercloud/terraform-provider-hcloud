package server_test

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestServerResource_Basic(t *testing.T) {
	var s hcloud.Server

	res := &server.RData{
		Name:  "server-basic",
		Type:  "cx11",
		Image: "ubuntu-20.04",
	}
	res.SetRName("server-basic")
	resRenamed := &server.RData{Name: res.Name + "-renamed", Type: res.Type, Image: res.Image}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_server", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-basic--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
				),
			},
			{
				// Try to import the newly created Server
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ssh_keys", "user_data", "keep_disk"},
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields and renaming the Server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("server-basic-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "image", res.Image),
				),
			},
		},
	})
}

func TestServerResource_Resize(t *testing.T) {
	var s hcloud.Server

	res := &server.RData{
		Name:  "server-resize",
		Type:  "cx11",
		Image: "ubuntu-20.04",
	}
	res.SetRName("server-resize")
	resResized := &server.RData{Name: res.Name, Type: "cx21", Image: res.Image, KeepDisk: true}
	resResized.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_server", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
				),
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields and renaming the Server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resResized,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resResized.TFID(), "name",
						fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resResized.TFID(), "server_type", resResized.Type),
					resource.TestCheckResourceAttr(resResized.TFID(), "image", res.Image),
				),
			},
		},
	})
}

func TestServerResource_ChangeUserData(t *testing.T) {
	var s, s2 hcloud.Server

	res := &server.RData{
		Name:     "server-userdata",
		Type:     "cx11",
		Image:    "ubuntu-20.04",
		UserData: "stuff",
	}
	res.SetRName("server-userdata")
	resChangedUserdata := &server.RData{Name: res.Name, Type: res.Type, Image: res.Image, UserData: "updated stuff"}
	resChangedUserdata.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_server", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(res.TFID(), "user_data", userDataHashSum(res.UserData)),
				),
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields and renaming the Server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resChangedUserdata,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s2)),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "name",
						fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "user_data", userDataHashSum(resChangedUserdata.UserData)),
					testsupport.LiftTCF(isRecreated(&s2, &s)),
				),
			},
		},
	})
}

func TestServerResource_ISO(t *testing.T) {
	var s hcloud.Server

	res := &server.RData{
		Name:     "server-iso",
		Type:     "cx11",
		Image:    "ubuntu-20.04",
		UserData: "stuff",
		ISO:      "3500",
	}
	res.SetRName("server-iso")
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_server", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-iso--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(res.TFID(), "iso", res.ISO),
				),
			},
		},
	})
}

func isRecreated(new, old *hcloud.Server) func() error {
	return func() error {
		if new.ID == old.ID {
			return fmt.Errorf("new server is the same as server cert %d", old.ID)
		}
		return nil
	}
}

func userDataHashSum(userData string) string {
	sum := sha1.Sum([]byte(userData))
	return base64.StdEncoding.EncodeToString(sum[:])
}
