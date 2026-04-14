package primaryip_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccPrimaryIPResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcPrimaryIP hcloud.PrimaryIP

	res1 := &primaryip.RData{
		Name:             "primary-ip",
		Type:             "ipv6",
		Location:         teste2e.TestLocationName,
		AutoDelete:       false,
		DeleteProtection: true,
		Labels:           map[string]string{"key": "value"},
	}
	res1.SetRName("main")

	res3 := testtemplate.DeepCopy(t, res1)
	res3.Name = res1.Name + "-changed"
	res3.Labels = map[string]string{"key": "changed"}
	res3.AutoDelete = true
	res3.DeleteProtection = false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &hcPrimaryIP)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res1,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res3.TFID(), plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("primary-ip--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(res1.Type)),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("value")})),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("ip_address"), testsupport.StringExactFromFunc(func() string { return hcPrimaryIP.IP.String() })),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("auto_delete"), knownvalue.Bool(false)),
				},
			},
			{
				ResourceName:            res1.TFID(),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"datacenter"},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res3,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Make sure that it's actually an update and not a replacement
						plancheck.ExpectResourceAction(res3.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("primary-ip-changed--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("type"), knownvalue.StringExact(res3.Type)),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("labels"), knownvalue.MapExact(map[string]knownvalue.Check{"key": knownvalue.StringExact("changed")})),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("ip_address"), testsupport.StringExactFromFunc(func() string { return hcPrimaryIP.IP.String() })),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("auto_delete"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestAccPrimaryIPResource_ConfigValidation(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcPrimaryIP hcloud.PrimaryIP

	res1 := &primaryip.RData{
		Name:       "primary-ip",
		Type:       "ipv6",
		Location:   teste2e.TestLocationName,
		AssigneeID: "1",
		// Test missing assignee_type
	}
	res1.SetRName("main")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &hcPrimaryIP)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res1,
				),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`These attributes must be configured together: \[assignee_id,assignee_type\]`),
			},
		},
	})
}

func TestAccPrimaryIPResource_WithServer(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcServer     hcloud.Server
		hcPrimaryIPA hcloud.PrimaryIP
		hcPrimaryIPB hcloud.PrimaryIP
		hcPrimaryIPC hcloud.PrimaryIP
	)

	// Step 1
	res1A := &primaryip.RData{
		Name:       "a",
		Type:       "ipv4",
		Location:   teste2e.TestLocationName,
		AutoDelete: false,
	}
	res1A.SetRName("a")

	res1B := &primaryip.RData{
		Name:       "b",
		Type:       "ipv6",
		Location:   teste2e.TestLocationName,
		AutoDelete: false,
	}
	res1B.SetRName("b")

	res1C := &primaryip.RData{
		Name:       "c",
		Type:       "ipv4",
		Location:   teste2e.TestLocationName,
		AutoDelete: false,
	}
	res1C.SetRName("c")

	res1Server := &server.RData{
		Name:         "primary-ip",
		Type:         teste2e.TestServerType,
		Image:        teste2e.TestImage,
		LocationName: teste2e.TestLocationName,
		PublicNet: map[string]any{
			"ipv4_enabled": true,
			"ipv6_enabled": true,
			"ipv4":         res1A.TFID() + ".id",
			"ipv6":         res1B.TFID() + ".id",
		},
	}
	res1Server.SetRName("primary_ip")

	// Step 2
	res2A := testtemplate.DeepCopy(t, res1A)
	res2B := testtemplate.DeepCopy(t, res1B)
	res2C := testtemplate.DeepCopy(t, res1C)

	res2Server := testtemplate.DeepCopy(t, res1Server)
	res2Server.PublicNet = map[string]any{
		"ipv4":         res2C.TFID() + ".id",
		"ipv6_enabled": false,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &hcServer)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &hcPrimaryIPA)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &hcPrimaryIPB)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &hcPrimaryIPC)),
		),
		Steps: []resource.TestStep{
			{
				// Create a new primary ip & server using the required values only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res1A,
					"testdata/r/hcloud_primary_ip", res1B,
					"testdata/r/hcloud_primary_ip", res1C,
					"testdata/r/hcloud_server", res1Server,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1A.TFID(), primaryip.ByID(t, &hcPrimaryIPA)),
					testsupport.CheckResourceExists(res1B.TFID(), primaryip.ByID(t, &hcPrimaryIPB)),
					testsupport.CheckResourceExists(res1C.TFID(), primaryip.ByID(t, &hcPrimaryIPC)),
					testsupport.CheckResourceExists(res1Server.TFID(), server.ByID(t, &hcServer)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1A.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("a--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res1B.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("b--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res1C.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("c--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res1A.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					statecheck.ExpectKnownValue(res1B.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(res1C.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					// Because the primary ips were created before the server, the
					// assignee_id is not refreshed after being attached to the server.
					statecheck.ExpectKnownValue(res1A.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(res1B.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(res1C.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(res1A.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
					statecheck.ExpectKnownValue(res1B.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
					statecheck.ExpectKnownValue(res1C.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
				},
			},
			{
				// Noop step to check the state after it was refreshed
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res1A,
					"testdata/r/hcloud_primary_ip", res1B,
					"testdata/r/hcloud_primary_ip", res1C,
					"testdata/r/hcloud_server", res1Server,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res1A.TFID(), plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(res1B.TFID(), plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(res1C.TFID(), plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(res1Server.TFID(), plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1A.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServer.ID })),
					statecheck.ExpectKnownValue(res1B.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServer.ID })),
					statecheck.ExpectKnownValue(res1C.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return 0 })),
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res2A,
					"testdata/r/hcloud_primary_ip", res2B,
					"testdata/r/hcloud_primary_ip", res2C,
					"testdata/r/hcloud_server", res2Server,
				),
				Check: resource.ComposeTestCheckFunc(
					// assign current hcloud primary ips + new server to local variables + check its existence
					testsupport.CheckResourceExists(res2A.TFID(), primaryip.ByID(t, &hcPrimaryIPA)),
					testsupport.CheckResourceExists(res2B.TFID(), primaryip.ByID(t, &hcPrimaryIPB)),
					testsupport.CheckResourceExists(res2C.TFID(), primaryip.ByID(t, &hcPrimaryIPC)),
					testsupport.CheckResourceExists(res2Server.TFID(), server.ByID(t, &hcServer)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res2A.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("a--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res2B.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("b--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res2C.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("c--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(res2A.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					statecheck.ExpectKnownValue(res2B.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(res2C.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					statecheck.ExpectKnownValue(res2A.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
					statecheck.ExpectKnownValue(res2B.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
					statecheck.ExpectKnownValue(res2C.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
				},
			},
			{
				// Noop step to check the state after it was refreshed
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res2A,
					"testdata/r/hcloud_primary_ip", res2B,
					"testdata/r/hcloud_primary_ip", res2C,
					"testdata/r/hcloud_server", res2Server,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2A.TFID(), plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(res2B.TFID(), plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(res2C.TFID(), plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction(res2Server.TFID(), plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res2A.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return 0 })),
					statecheck.ExpectKnownValue(res2B.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return 0 })),
					statecheck.ExpectKnownValue(res2C.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServer.ID })),
				},
			},
		},
	})
}

func TestAccPrimaryIPResource_Reassign(t *testing.T) {
	var (
		hcServerA   hcloud.Server
		hcServerB   hcloud.Server
		hcPrimaryIP hcloud.PrimaryIP
	)

	// Step 1
	res1ServerA := &server.RData{
		Name:         "a",
		Type:         teste2e.TestServerType,
		Image:        teste2e.TestImage,
		LocationName: teste2e.TestLocationName,
		PublicNet: map[string]any{
			"ipv4_enabled": false,
			"ipv6_enabled": true,
		},
	}
	res1ServerA.SetRName("a")

	res1ServerB := testtemplate.DeepCopy(t, res1ServerA)
	res1ServerB.SetRName("b")
	res1ServerB.Name = "b"

	// Step 2
	res2ServerA := testtemplate.DeepCopy(t, res1ServerA)
	res2ServerB := testtemplate.DeepCopy(t, res1ServerB)

	res2 := &primaryip.RData{
		Name:         "primary-ip",
		Type:         "ipv4",
		AssigneeID:   res1ServerA.TFID() + ".id",
		AssigneeType: "server",
		AutoDelete:   false,
	}
	res2.SetRName("main")

	// Step 3
	res3ServerA := testtemplate.DeepCopy(t, res2ServerA)
	res3ServerB := testtemplate.DeepCopy(t, res2ServerB)

	res3 := testtemplate.DeepCopy(t, res2)
	res3.AssigneeID = res3ServerB.TFID() + ".id"

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &hcServerA)),
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &hcServerB)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &hcPrimaryIP)),
		),
		Steps: []resource.TestStep{
			{
				// Create two servers and shut them down to freely assign/unassign primary ip
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res1ServerA,
					"testdata/r/hcloud_server", res1ServerB,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1ServerA.TFID(), server.ByID(t, &hcServerA)),
					testsupport.CheckResourceExists(res1ServerB.TFID(), server.ByID(t, &hcServerB)),
				),
				PostApplyFunc: func() {
					client, err := testsupport.CreateClient()
					if err != nil {
						t.Fatalf("Error in PostApplyFunc: %v", err)
					}
					actionOne, _, err := client.Server.Poweroff(t.Context(), &hcServerA)
					if err != nil {
						t.Fatalf("Error in PostApplyFunc: %v", err)
					}
					actionTwo, _, err := client.Server.Poweroff(t.Context(), &hcServerB)
					if err != nil {
						t.Fatalf("Error in PostApplyFunc: %v", err)
					}

					err = client.Action.WaitFor(t.Context(), actionOne, actionTwo)
					if err != nil {
						t.Fatalf("Error in PostApplyFunc: %v", err)
					}
				},
			},
			{
				// Create primary IP and assign it to the first server
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res2ServerA,
					"testdata/r/hcloud_server", res2ServerB,
					"testdata/r/hcloud_primary_ip", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res2ServerA.TFID(), server.ByID(t, &hcServerA)),
					testsupport.CheckResourceExists(res2ServerB.TFID(), server.ByID(t, &hcServerB)),
					testsupport.CheckResourceExists(res2.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res2.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerA.ID })),
				},
			},
			{
				// Reassign IP to second server
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res3ServerA,
					"testdata/r/hcloud_server", res3ServerB,
					"testdata/r/hcloud_primary_ip", res3,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res3ServerA.TFID(), server.ByID(t, &hcServerA)),
					testsupport.CheckResourceExists(res3ServerB.TFID(), server.ByID(t, &hcServerB)),
					testsupport.CheckResourceExists(res3.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res3.TFID(), tfjsonpath.New("assignee_id"), testsupport.Int64ExactFromFunc(func() int64 { return hcServerB.ID })),
				},
			},
		},
	})
}

func TestAccPrimaryIPResource_DeleteProtection(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcPrimaryIP hcloud.PrimaryIP

	unprotected := &primaryip.RData{
		Name:             "main",
		Type:             "ipv6",
		Location:         teste2e.TestLocationName,
		DeleteProtection: false,
	}
	unprotected.SetRName("main")

	protected := testtemplate.DeepCopy(t, unprotected)
	protected.DeleteProtection = true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &hcPrimaryIP)),
		Steps: []resource.TestStep{
			{
				// Create protected primary IP.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", protected),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(protected.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(protected.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(protected.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(true)),
				},
			},
			{
				// Delete protected primary IP.
				Config:      tmplMan.Render(t, "testdata/r/hcloud_primary_ip", protected),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Error code: protected`),
			},
			{
				// Change primary IP protection.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", unprotected),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(unprotected.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(protected.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(protected.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
				},
			},
			{
				// Delete unprotected primary IP.
				Config:  tmplMan.Render(t, "testdata/r/hcloud_primary_ip", unprotected),
				Destroy: true,
			},
		},
	})
}

func TestAccPrimaryIPResource_DatacenterToLocation(t *testing.T) {
	// Test for the "datacenter" deprecation, to make sure that its possible to move to "location" attribute
	// See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters
	tmplMan := testtemplate.Manager{}

	res1 := &primaryip.RData{
		Name:       "datacenter-to-location",
		Type:       "ipv6",
		Datacenter: teste2e.TestDataCenter,
	}
	res1.SetRName("main")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.Datacenter = ""
	res2.Location = teste2e.TestLocationName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create primary IP in Datacenter.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", res1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
			{
				// Change config to Location.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", res2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
		},
	})
}

func TestAccPrimaryIPResource_DatacenterToLocationForceNew(t *testing.T) {
	// Test for the "datacenter" deprecation, to make sure that its possible to move to "location" attribute
	// See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters
	tmplMan := testtemplate.Manager{}

	res1 := &primaryip.RData{
		Name:         "datacenter-to-location",
		Type:         "ipv6",
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server", // Attribute was still required in previous versions
	}
	res1.SetRName("main")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.Datacenter = ""
	res2.Location = teste2e.TestLocationName
	res2.AssigneeType = ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.57.0",
						Source:            "hetznercloud/hcloud",
					},
				},
				// Create primary IP in Datacenter.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", res1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
				},
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				// Change config to Location.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", res2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res2.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(res2.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
		},
	})
}

func TestAccPrimaryIPResource_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &primaryip.RData{
		Name:         "main",
		Type:         "ipv6",
		Location:     teste2e.TestLocationName,
		AssigneeType: "server",
		// Labels will default {} after the upgrade, this is a workaround to make the tests pass
		Labels: map[string]string{"key": "value"},
	}
	res.SetRName("main")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.60.1",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
				),
				PlanOnly: true,
			},
		},
	})
}
