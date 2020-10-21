package floatingip_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestFloatingIPAssignmentResource_Basic(t *testing.T) {
	var s hcloud.Server
	var s2 hcloud.Server
	var f hcloud.FloatingIP
	tmplMan := testtemplate.Manager{}
	resServer := &server.RData{
		Name:  "fip-assignment",
		Type:  "cx11",
		Image: "ubuntu-20.04",
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-fip-assignment-%d", tmplMan.RandInt),
		},
	}
	resServer.SetRName("server_assignment")

	resServer2 := &server.RData{
		Name:  "fip-assignment-2",
		Type:  "cx11",
		Image: "ubuntu-20.04",
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-fip-assignment-%d", tmplMan.RandInt),
		},
	}
	resServer2.SetRName("server2_assignment")

	resFloatingIP := &floatingip.RData{
		Name:             "fip-assignment",
		Type:             "ipv4",
		HomeLocationName: "fsn1",
	}
	resFloatingIP.SetRName("floating_ip_assignment")

	res := &floatingip.RDataAssignment{
		FloatingIPID: resFloatingIP.TFID() + ".id",
		ServerID:     resServer.TFID() + ".id",
	}

	resMove := &floatingip.RDataAssignment{
		FloatingIPID: resFloatingIP.TFID() + ".id",
		ServerID:     resServer2.TFID() + ".id",
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new RDNS using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server", resServer2,
					"testdata/r/hcloud_floating_ip", resFloatingIP,
					"testdata/r/hcloud_floating_ip_assignment", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer.TFID(), server.ByID(t, &s)),
					testsupport.CheckResourceExists(resFloatingIP.TFID(), floatingip.ByID(t, &f)),
				),
			},
			{
				// Try to import the newly created Floating IP assignment
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%d", f.ID), nil
				},
			},
			{
				// Move the floating IP to another server
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_server", resServer2,
					"testdata/r/hcloud_floating_ip", resFloatingIP,
					"testdata/r/hcloud_floating_ip_assignment", resMove,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resServer2.TFID(), server.ByID(t, &s2)),
					testsupport.CheckResourceExists(resFloatingIP.TFID(), floatingip.ByID(t, &f)),
				),
			},
		},
	})
}
