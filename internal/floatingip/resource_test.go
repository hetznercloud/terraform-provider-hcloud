package floatingip_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccFloatingIPResource(t *testing.T) {
	var fip hcloud.FloatingIP

	res := &floatingip.RData{
		Name:             "floatingip-test",
		Type:             "ipv4",
		Labels:           nil,
		HomeLocationName: teste2e.TestLocationName,
	}
	resRenamed := &floatingip.RData{Name: res.Name + "-renamed", Type: res.Type, HomeLocationName: res.HomeLocationName}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(floatingip.ResourceType, floatingip.ByID(t, &fip)),
		Steps: []resource.TestStep{
			{
				// Create a new Floating IP using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_floating_ip", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), floatingip.ByID(t, &fip)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("floatingip-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "type", res.Type),
				),
			},
			{
				// Try to import the newly created Floating IP
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the Floating IP created in the previous step by
				// setting all optional fields and renaming the Floating IP.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("floatingip-test-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "type", res.Type),
				),
			},
		},
	})
}
func TestAccFloatingIPResource_WithServer(t *testing.T) {
	var fip hcloud.FloatingIP
	tmplMan := testtemplate.Manager{}

	resSSHKey := sshkey.NewRData(t, "server-floating-ip-withserver")
	resServer := &server.RData{
		Name:  "floating-ip-test",
		Type:  teste2e.TestServerType,
		Image: teste2e.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-fip-assignment-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	resServer.SetRName("server_assignment")

	res := &floatingip.RData{
		Name:     "floatingip-server-test",
		Type:     "ipv4",
		Labels:   nil,
		ServerID: resServer.TFID() + ".id",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(floatingip.ResourceType, floatingip.ByID(t, &fip)),
		Steps: []resource.TestStep{
			{
				// Create a new Floating IP using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_floating_ip", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), floatingip.ByID(t, &fip)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("floatingip-server-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "type", res.Type),
				),
			},
			{
				// Try to import the newly created Floating IP
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFloatingIPResource_Protection(t *testing.T) {
	var (
		fip hcloud.FloatingIP

		res = &floatingip.RData{
			Name:             "floatingip-protection",
			Type:             "ipv4",
			Labels:           nil,
			HomeLocationName: teste2e.TestLocationName,
			DeleteProtection: true,
		}

		updateProtection = func(d *floatingip.RData, protection bool) *floatingip.RData {
			d.DeleteProtection = protection
			return d
		}
	)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(floatingip.ResourceType, floatingip.ByID(t, &fip)),
		Steps: []resource.TestStep{
			{
				// Create a new Floating IP using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), floatingip.ByID(t, &fip)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("floatingip-protection--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
				),
			},
			{
				// Update delete protection
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", updateProtection(res, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
				),
			},
		},
	})
}
