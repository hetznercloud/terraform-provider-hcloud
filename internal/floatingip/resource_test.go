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
		Labels:           map[string]string{"key": "value"},
	}
	resA1.SetRName("ipv6")

	resB1 := &floatingip.RData{
		Name:             "floating-ipv4",
		Type:             "ipv4",
		HomeLocationName: teste2e.TestLocationName,
		Labels:           map[string]string{"key": "value"},
	}
	resB1.SetRName("ipv4")

	resA3 := testtemplate.DeepCopy(t, resA1)
	resA3.Name = resA1.Name + "-changed"
	resA3.Labels = map[string]string{"key": "changed"}

	resB3 := testtemplate.DeepCopy(t, resB1)
	resB3.Name = resB1.Name + "-changed"
	resB3.Labels = map[string]string{"key": "changed"}

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
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),

					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resB1.Type)),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("ip_address"), testsupport.StringExactFromFunc(func() string { return hcFloatingIPB.IP.String() })),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("ip_network"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
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

func TestAccFloatingIPResource_Protection(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcFloatingIPA hcloud.FloatingIP

	resA1 := &floatingip.RData{
		Name:             "floating-ipv6",
		Type:             "ipv6",
		HomeLocationName: teste2e.TestLocationName,
		DeleteProtection: true,
	}
	resA1.SetRName("ipv6")

	resA3 := testtemplate.DeepCopy(t, resA1)
	resA3.DeleteProtection = false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(floatingip.ResourceType, floatingip.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", resA1,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resA1.TFID(), plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resA1.TFID(), testsupport.CopyAPIResource(&hcFloatingIPA, floatingip.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resA1.Type)),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(true)),
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
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure that it's actually an update and not a replacement
						plancheck.ExpectResourceAction(resA3.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resA3.Type)),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccFloatingIPResource_Server(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcFloatingIPA hcloud.FloatingIP
	var hcFloatingIPB hcloud.FloatingIP
	var hcServerA hcloud.Server
	var hcServerB hcloud.Server

	srvs := server.NewBlueprint(t)

	resA1 := &floatingip.RData{
		Name:             "floating-ipv6",
		Type:             "ipv6",
		HomeLocationName: teste2e.TestLocationName,
		ServerID:         srvs.ServerA.TFID() + ".id",
	}
	resA1.SetRName("ipv6")

	resB1 := &floatingip.RData{
		Name:             "floating-ipv4",
		Type:             "ipv4",
		HomeLocationName: teste2e.TestLocationName,
		ServerID:         srvs.ServerA.TFID() + ".id",
	}
	resB1.SetRName("ipv4")

	// Change Server ID
	resA3 := testtemplate.DeepCopy(t, resA1)
	resA3.ServerID = srvs.ServerB.TFID() + ".id"

	resB3 := testtemplate.DeepCopy(t, resB1)
	resB3.ServerID = srvs.ServerB.TFID() + ".id"

	// Remove Server ID
	resA4 := testtemplate.DeepCopy(t, resA3)
	resA4.ServerID = ""

	resB4 := testtemplate.DeepCopy(t, resB3)
	resB4.ServerID = ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(floatingip.ResourceType, floatingip.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_server", srvs.ServerB,
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
					testsupport.CheckAPIResourcePresent(srvs.ServerA.TFID(), testsupport.CopyAPIResource(&hcServerA, server.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(srvs.ServerB.TFID(), testsupport.CopyAPIResource(&hcServerB, server.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resA1.Type)),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerA.ID })),

					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resB1.Type)),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerA.ID })),
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
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_server", srvs.ServerB,
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
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resA3.Type)),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerB.ID })),

					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resB3.Type)),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerB.ID })),
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_server", srvs.ServerB,
					"testdata/r/hcloud_floating_ip", resA4,
					"testdata/r/hcloud_floating_ip", resB4,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure that it's actually an update and not a replacement
						plancheck.ExpectResourceAction(resA4.TFID(), plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(resB4.TFID(), plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA4.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resA4.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resA4.Type)),
					statecheck.ExpectKnownValue(resA4.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerB.ID })),

					statecheck.ExpectKnownValue(resB4.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("floating-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(resB4.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(resB4.Type)),
					statecheck.ExpectKnownValue(resB4.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerB.ID })),
				},
			},
		},
	})
}
