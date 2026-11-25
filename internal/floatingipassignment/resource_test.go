package floatingipassignment_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingipassignment"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccFloatingIPResourceAssignment(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcFloatingIPA hcloud.FloatingIP
	var hcFloatingIPB hcloud.FloatingIP
	var hcServerA hcloud.Server
	var hcServerB hcloud.Server

	srvs := server.NewBlueprint(t)

	resFloatingIPA := &floatingip.RData{
		Name:             "floating-ipv6",
		Type:             "ipv6",
		HomeLocationName: teste2e.TestLocationName,
	}
	resFloatingIPA.SetRName("ipv6")

	resFloatingIPB := &floatingip.RData{
		Name:             "floating-ipv4",
		Type:             "ipv4",
		HomeLocationName: teste2e.TestLocationName,
	}
	resFloatingIPB.SetRName("ipv4")

	resA1 := &floatingipassignment.RData{
		FloatingIPID: resFloatingIPA.TFID() + ".id",
		ServerID:     srvs.ServerA.TFID() + ".id",
	}
	resA1.SetRName("ipv6")

	resB1 := &floatingipassignment.RData{
		FloatingIPID: resFloatingIPB.TFID() + ".id",
		ServerID:     srvs.ServerA.TFID() + ".id",
	}
	resB1.SetRName("ipv4")

	resA3 := testtemplate.DeepCopy(t, resA1)
	resA3.ServerID = srvs.ServerB.TFID() + ".id"

	resB3 := testtemplate.DeepCopy(t, resB1)
	resB3.ServerID = srvs.ServerB.TFID() + ".id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(floatingip.ResourceType, floatingip.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_server", srvs.ServerB,
					"testdata/r/hcloud_floating_ip", resFloatingIPA,
					"testdata/r/hcloud_floating_ip", resFloatingIPB,
					"testdata/r/hcloud_floating_ip_assignment", resA1,
					"testdata/r/hcloud_floating_ip_assignment", resB1,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resA1.TFID(), plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resB1.TFID(), plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resFloatingIPA.TFID(), testsupport.CopyAPIResource(&hcFloatingIPA, floatingip.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resFloatingIPB.TFID(), testsupport.CopyAPIResource(&hcFloatingIPB, floatingip.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(srvs.ServerA.TFID(), testsupport.CopyAPIResource(&hcServerA, server.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(srvs.ServerB.TFID(), testsupport.CopyAPIResource(&hcServerB, server.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("floating_ip_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcFloatingIPA.ID })),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerA.ID })),

					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("floating_ip_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcFloatingIPB.ID })),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerA.ID })),
				},
			},
			{
				ResourceName:      resA1.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvs.ServerA,
					"testdata/r/hcloud_server", srvs.ServerB,
					"testdata/r/hcloud_floating_ip", resFloatingIPA,
					"testdata/r/hcloud_floating_ip", resFloatingIPB,
					"testdata/r/hcloud_floating_ip_assignment", resA3,
					"testdata/r/hcloud_floating_ip_assignment", resB3,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resA3.TFID(), plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction(resB3.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(resFloatingIPA.TFID(), testsupport.CopyAPIResource(&hcFloatingIPA, floatingip.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(resFloatingIPB.TFID(), testsupport.CopyAPIResource(&hcFloatingIPB, floatingip.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(srvs.ServerA.TFID(), testsupport.CopyAPIResource(&hcServerA, server.GetAPIResource())),
					testsupport.CheckAPIResourcePresent(srvs.ServerB.TFID(), testsupport.CopyAPIResource(&hcServerB, server.GetAPIResource())),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("floating_ip_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcFloatingIPA.ID })),
					statecheck.ExpectKnownValue(resA3.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerB.ID })),

					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("floating_ip_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcFloatingIPB.ID })),
					statecheck.ExpectKnownValue(resB3.TFID(), tfjsonpath.New("server_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerB.ID })),
				},
			},
		},
	})
}
