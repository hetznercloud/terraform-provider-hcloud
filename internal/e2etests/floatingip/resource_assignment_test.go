package floatingip_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"

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

	resSSHKey := sshkey.NewRData(t, "server-floating-ip-basic")
	resServer := &server.RData{
		Name:  "fip-assignment",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-fip-assignment-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("server_assignment")

	resServer2 := &server.RData{
		Name:  "fip-assignment-2",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-fip-assignment-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer2.SetRName("server2_assignment")

	resFloatingIP := &floatingip.RData{
		Name:             "fip-assignment",
		Type:             "ipv4",
		HomeLocationName: e2etests.TestLocationName,
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new RDNS using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
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
					"testdata/r/hcloud_ssh_key", resSSHKey,
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
