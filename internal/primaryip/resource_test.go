package primaryip_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

func TestAccPrimaryIPResource(t *testing.T) {
	var pip hcloud.PrimaryIP

	res := &primaryip.RData{
		Name:         "primaryip-test",
		Type:         "ipv4",
		Labels:       nil,
		Location:     teste2e.TestLocationName,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	resRenamed := &primaryip.RData{
		Name:         res.Name + "-renamed",
		Type:         res.Type,
		AssigneeType: res.AssigneeType,
		Location:     res.Location,
		AutoDelete:   res.AutoDelete,
	}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &pip)),
		Steps: []resource.TestStep{
			{
				// Create a new primary IP using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), primaryip.ByID(t, &pip)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("primaryip-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "type", res.Type),
				),
			},
			{
				// Try to import the newly created primary IP
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the primary IP created in the previous step by
				// setting all optional fields and renaming the primary IP.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("primaryip-test-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "type", res.Type),
				),
			},
		},
	})
}

func TestAccPrimaryIPResource_WithServer(t *testing.T) {
	var srv hcloud.Server
	var primaryIPv4One hcloud.PrimaryIP
	var primaryIPv4Two hcloud.PrimaryIP
	var primaryIPv6One hcloud.PrimaryIP
	primaryIPv4OneRes := &primaryip.RData{
		Name:         "primaryip-test-v4-one",
		Type:         "ipv4",
		Labels:       nil,
		Location:     teste2e.TestLocationName,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPv4OneRes.SetRName("primary_ip_v4_test")

	primaryIPv6OneRes := &primaryip.RData{
		Name:         "primaryip-test-v6-one",
		Type:         "ipv6",
		Labels:       nil,
		Location:     teste2e.TestLocationName,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPv6OneRes.SetRName("primary_ip_v6_test")

	primaryIPv4TwoRes := &primaryip.RData{
		Name:         "primaryip-test-v4-two",
		Type:         "ipv4",
		Labels:       nil,
		Location:     teste2e.TestLocationName,
		AssigneeType: "server",
		AutoDelete:   true,
	}
	primaryIPv4TwoRes.SetRName("primary_ip_v4_two_test")

	testServerRes := &server.RData{
		Name:         "server-test",
		Type:         teste2e.TestServerType,
		Image:        teste2e.TestImage,
		LocationName: teste2e.TestLocationName,
		Labels:       nil,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": true,
			"ipv6_enabled": true,
			"ipv4":         primaryIPv4OneRes.TFID() + ".id",
			"ipv6":         primaryIPv6OneRes.TFID() + ".id",
		},
	}

	testServerUpdatedRes := &server.RData{
		Name:         testServerRes.Name,
		Type:         testServerRes.Type,
		Image:        testServerRes.Image,
		LocationName: testServerRes.LocationName,
		Labels:       testServerRes.Labels,
		PublicNet: map[string]interface{}{
			"ipv4":         primaryIPv4TwoRes.TFID() + ".id",
			"ipv6_enabled": false,
		},
	}
	testServerUpdatedRes.SetRName(testServerRes.RName())

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &primaryIPv4One)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &primaryIPv4Two)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &primaryIPv6One)),
		),
		Steps: []resource.TestStep{
			{
				// Create a new primary ip & server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", primaryIPv4OneRes,
					"testdata/r/hcloud_primary_ip", primaryIPv6OneRes,
					"testdata/r/hcloud_primary_ip", primaryIPv4TwoRes,
					"testdata/r/hcloud_server", testServerRes),

				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(primaryIPv4OneRes.TFID(), primaryip.ByID(t, &primaryIPv4One)),
					testsupport.CheckResourceExists(primaryIPv4TwoRes.TFID(), primaryip.ByID(t, &primaryIPv4Two)),
					testsupport.CheckResourceExists(primaryIPv6OneRes.TFID(), primaryip.ByID(t, &primaryIPv6One)),
					testsupport.CheckResourceExists(testServerRes.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(primaryIPv4OneRes.TFID(), "name",
						fmt.Sprintf("primaryip-test-v4-one--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(primaryIPv6OneRes.TFID(), "name",
						fmt.Sprintf("primaryip-test-v6-one--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(testServerRes.TFID(), "name",
						fmt.Sprintf("server-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(primaryIPv4OneRes.TFID(), "type", primaryIPv4OneRes.Type),
					resource.TestCheckResourceAttr(testServerRes.TFID(), "server_type", testServerRes.Type),
					resource.TestCheckResourceAttr(primaryIPv4OneRes.TFID(), "assignee_id", util.FormatID(primaryIPv4One.ID)),
				),
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", primaryIPv4OneRes,
					"testdata/r/hcloud_primary_ip", primaryIPv6OneRes,
					"testdata/r/hcloud_primary_ip", primaryIPv4TwoRes,
					"testdata/r/hcloud_server", testServerUpdatedRes),
				Check: resource.ComposeTestCheckFunc(
					// assign current hcloud primary ips + new server to local variables + check its existence
					testsupport.CheckResourceExists(primaryIPv4OneRes.TFID(), primaryip.ByID(t, &primaryIPv4One)),
					testsupport.CheckResourceExists(primaryIPv4TwoRes.TFID(), primaryip.ByID(t, &primaryIPv4Two)),
					testsupport.CheckResourceExists(primaryIPv6OneRes.TFID(), primaryip.ByID(t, &primaryIPv6One)),
					testsupport.CheckResourceExists(testServerUpdatedRes.TFID(), server.ByID(t, &srv)),
					func(_ *terraform.State) error {
						// check current hcloud state, validating if ips got assigned / unassigned correctly
						if primaryIPv4Two.AssigneeID == srv.ID &&
							primaryIPv6One.AssigneeID != srv.ID &&
							primaryIPv4One.AssigneeID != srv.ID {
							return nil
						}
						// nolint:revive
						return fmt.Errorf(`state is not as expected:
primary IP v4 two has assignee id %d which not equals target server id %d
primary IP v4 one has assignee id %d and should shouldnt be assigned to server id %d
primary IP v6 one has assignee id %d and should shouldnt be assigned to server id %d
`,
							primaryIPv4Two.AssigneeID, srv.ID,
							primaryIPv4One.AssigneeID, srv.ID,
							primaryIPv6One.AssigneeID, srv.ID,
						)
					}),
			},
		},
	})
}

func TestAccPrimaryIPResource_Reassign(t *testing.T) {
	var srvOne hcloud.Server
	var srvTwo hcloud.Server
	var primaryIP hcloud.PrimaryIP

	testServerOneRes := &server.RData{
		Name:         "primary-ip-reassign-one",
		Type:         teste2e.TestServerType,
		Image:        teste2e.TestImage,
		LocationName: teste2e.TestLocationName,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": false,
			"ipv6_enabled": true,
		},
	}
	testServerOneRes.SetRName("one")

	testServerTwoRes := testtemplate.DeepCopy(t, testServerOneRes)
	testServerTwoRes.Name = "primary-ip-reassign-two"
	testServerTwoRes.SetRName("two")

	initialIPRes := &primaryip.RData{
		Name:         "primaryip-test-reassign",
		Type:         "ipv4",
		Labels:       nil,
		AssigneeType: "server",
		AssigneeID:   testServerOneRes.TFID() + ".id",
		AutoDelete:   false,
	}
	initialIPRes.SetRName("reassign")

	reassignedIPRes := testtemplate.DeepCopy(t, initialIPRes)
	reassignedIPRes.AssigneeID = testServerTwoRes.TFID() + ".id"

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srvOne)),
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srvTwo)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &primaryIP)),
		),
		Steps: []resource.TestStep{
			{
				// Create two servers and shut them down to freely assign/unassign primary ip
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", testServerOneRes,
					"testdata/r/hcloud_server", testServerTwoRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(testServerOneRes.TFID(), server.ByID(t, &srvOne)),
					testsupport.CheckResourceExists(testServerTwoRes.TFID(), server.ByID(t, &srvTwo)),
				),
				PostApplyFunc: func() {
					client, err := testsupport.CreateClient()
					if err != nil {
						t.Fatalf("Error in PostApplyFunc: %v", err)
					}
					actionOne, _, err := client.Server.Poweroff(t.Context(), &srvOne)
					if err != nil {
						t.Fatalf("Error in PostApplyFunc: %v", err)
					}
					actionTwo, _, err := client.Server.Poweroff(t.Context(), &srvTwo)
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
					"testdata/r/hcloud_server", testServerOneRes,
					"testdata/r/hcloud_server", testServerTwoRes,
					"testdata/r/hcloud_primary_ip", initialIPRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(initialIPRes.TFID(), primaryip.ByID(t, &primaryIP)),
					resource.TestCheckResourceAttrWith(initialIPRes.TFID(), "assignee_id", func(value string) error {
						id, err := util.ParseID(value)
						if err != nil {
							return err
						}
						if id != srvOne.ID {
							return fmt.Errorf("wrong assignee_id, expected %d but got %d", srvOne.ID, id)
						}

						return nil
					}),
				),
			},
			{
				// Reassign IP to second server
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", testServerOneRes,
					"testdata/r/hcloud_server", testServerTwoRes,
					"testdata/r/hcloud_primary_ip", reassignedIPRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(reassignedIPRes.TFID(), "assignee_id", func(value string) error {
						id, err := util.ParseID(value)
						if err != nil {
							return err
						}
						if id != srvTwo.ID {
							return fmt.Errorf("wrong assignee_id, expected %d but got %d", srvTwo.ID, id)
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccPrimaryIPResource_FieldUpdates(t *testing.T) {
	var (
		pip hcloud.PrimaryIP

		res = &primaryip.RData{
			Name:             "primaryip-protection",
			Type:             "ipv4",
			Labels:           nil,
			Location:         teste2e.TestLocationName,
			AssigneeType:     "server",
			DeleteProtection: true,
			AutoDelete:       true,
		}

		updateFields = func(d *primaryip.RData, protection bool, autoDelete bool) *primaryip.RData {
			d.DeleteProtection = protection
			d.AutoDelete = autoDelete
			return d
		}
	)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &pip)),
		Steps: []resource.TestStep{
			{
				// Create a new primary IP using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), primaryip.ByID(t, &pip)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("primaryip-protection--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
				),
			},
			{
				// Update delete protection
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", updateFields(res, false, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
				),
			},
		},
	})
}

func TestAccPrimaryIPResource_DeleteProtection(t *testing.T) {
	var pip hcloud.PrimaryIP

	unprotected := &primaryip.RData{
		Name:             "primaryip-test",
		Type:             "ipv6",
		Location:         teste2e.TestLocationName,
		AssigneeType:     "server",
		DeleteProtection: false,
	}

	protected := &primaryip.RData{
		Name:             unprotected.Name,
		Type:             unprotected.Type,
		Location:         unprotected.Location,
		AssigneeType:     unprotected.AssigneeType,
		DeleteProtection: true,
	}

	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &pip)),
		Steps: []resource.TestStep{
			{
				// Create protected primary IP.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", protected),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(protected.TFID(), primaryip.ByID(t, &pip)),
					resource.TestCheckResourceAttr(protected.TFID(), "name", fmt.Sprintf("primaryip-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(protected.TFID(), "type", protected.Type),
					resource.TestCheckResourceAttr(protected.TFID(), "delete_protection", "true"),
				),
			},
			{
				// Delete protected primary IP.
				Config:      tmplMan.Render(t, "testdata/r/hcloud_primary_ip", protected),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Primary IP deletion is protected \(protected, .*\)`),
			},
			{
				// Change primary IP protection.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", unprotected),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(unprotected.TFID(), primaryip.ByID(t, &pip)),
					resource.TestCheckResourceAttr(unprotected.TFID(), "name", fmt.Sprintf("primaryip-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(unprotected.TFID(), "type", protected.Type),
					resource.TestCheckResourceAttr(unprotected.TFID(), "delete_protection", "false"),
				),
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

	resDC := &primaryip.RData{
		Name:         "primaryip-dc-to-location",
		Type:         "ipv6",
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
	}
	resDC.SetRName("dc_to_location")

	resLocation := testtemplate.DeepCopy(t, resDC)
	resLocation.Datacenter = ""
	resLocation.Location = teste2e.TestLocationName

	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create primary IP in Datacenter.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", resDC),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resDC.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(resDC.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
			{
				// Change config to Location.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", resLocation),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resDC.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(resDC.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
		},
	})
}

func TestAccPrimaryIPResource_DatacenterToLocationForceNew(t *testing.T) {
	// Test for the "datacenter" deprecation, to make sure that its possible to move to "location" attribute
	// See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters

	resDC := &primaryip.RData{
		Name:         "primaryip-dc-to-location",
		Type:         "ipv6",
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
	}
	resDC.SetRName("dc_to_location")

	resLocation := testtemplate.DeepCopy(t, resDC)
	resLocation.Datacenter = ""
	resLocation.Location = teste2e.TestLocationName

	tmplMan := testtemplate.Manager{}

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
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", resDC),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resDC.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
				},
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				// Change config to Location.
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", resLocation),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resLocation.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(resLocation.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
		},
	})
}
