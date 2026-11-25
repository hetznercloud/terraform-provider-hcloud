package floatingip_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccFloatingIPResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcFloatingIPA hcloud.FloatingIP
	var hcFloatingIPB hcloud.FloatingIP

	resA1 := &floatingip.RData{
		Name:             "floating-ipv6",
		Type:             "ipv6",
		HomeLocationName: teste2e.TestLocationName,
		DeleteProtection: true,
		Labels:           map[string]string{"key": "value"},
	}
	resA1.SetRName("ipv6")

	resB1 := &floatingip.RData{
		Name:             "floating-ipv4",
		Type:             "ipv4",
		HomeLocationName: teste2e.TestLocationName,
		DeleteProtection: true,
		Labels:           map[string]string{"key": "value"},
	}
	resB1.SetRName("ipv4")

	resA3 := testtemplate.DeepCopy(t, resA1)
	resA3.Name = resA1.Name + "-changed"
	resA3.Labels = map[string]string{"key": "changed"}
	resA3.DeleteProtection = false

	resB3 := testtemplate.DeepCopy(t, resB1)
	resB3.Name = resB1.Name + "-changed"
	resB3.Labels = map[string]string{"key": "changed"}
	resB3.DeleteProtection = false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(floatingip.ResourceType, floatingip.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resA1,
					"testdata/r/hcloud_floating_ip", resB1,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resA1.TFID(), plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resB1.TFID(), plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resA1.TFID(), testsupport.CopyAPIResource(&hcFloatingIPA, floatingip.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resB1.TFID(), testsupport.CopyAPIResource(&hcFloatingIPB, floatingip.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resA1.Type)),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("ip_address"), testsupport.StringExactFromFunc(func() string { return hcFloatingIPA.IP.String() })),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("ip_network"), testsupport.StringExactFromFunc(func() string { return hcFloatingIPA.Network.String() })),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),

					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resB1.Type)),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("ip_address"), testsupport.StringExactFromFunc(func() string { return hcFloatingIPB.IP.String() })),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("ip_network"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
				},
			},
			{
				ResourceName:            resA1.TFID(),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"datacenter"},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resA3,
					"testdata/r/hcloud_floating_ip", resB3,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure that it's actually an update and not a replacement
						plancheck.ExpectResourceAction(resA3.TFID(), plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction(resB3.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv6-changed--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resA3.Type)),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("changed")})),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("ip_address"), testsupport.StringExactFromFunc(func() string { return hcFloatingIPA.IP.String() })),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("ip_network"), testsupport.StringExactFromFunc(func() string { return hcFloatingIPA.Network.String() })),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),

					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv4-changed--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resB3.Type)),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("changed")})),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("ip_address"), testsupport.StringExactFromFunc(func() string { return hcFloatingIPB.IP.String() })),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("ip_network"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
				},
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
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
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
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
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
