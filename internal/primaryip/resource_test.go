package primaryip_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

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
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	resRenamed := &primaryip.RData{
		Name:         res.Name + "-renamed",
		Type:         res.Type,
		AssigneeType: res.AssigneeType,
		Datacenter:   res.Datacenter,
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
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPv4OneRes.SetRName("primary_ip_v4_test")

	primaryIPv6OneRes := &primaryip.RData{
		Name:         "primaryip-test-v6-one",
		Type:         "ipv6",
		Labels:       nil,
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPv6OneRes.SetRName("primary_ip_v6_test")

	primaryIPv4TwoRes := &primaryip.RData{
		Name:         "primaryip-test-v4-two",
		Type:         "ipv4",
		Labels:       nil,
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
		AutoDelete:   true,
	}
	primaryIPv4TwoRes.SetRName("primary_ip_v4_two_test")

	testServerRes := &server.RData{
		Name:       "server-test",
		Type:       teste2e.TestServerType,
		Image:      teste2e.TestImage,
		Datacenter: teste2e.TestDataCenter,
		Labels:     nil,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": true,
			"ipv6_enabled": true,
			"ipv4":         primaryIPv4OneRes.TFID() + ".id",
			"ipv6":         primaryIPv6OneRes.TFID() + ".id",
		},
	}

	testServerUpdatedRes := &server.RData{
		Name:       testServerRes.Name,
		Type:       testServerRes.Type,
		Image:      testServerRes.Image,
		Datacenter: testServerRes.Datacenter,
		Labels:     testServerRes.Labels,
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

func TestAccPrimaryIPResource_FieldUpdates(t *testing.T) {
	var (
		pip hcloud.PrimaryIP

		res = &primaryip.RData{
			Name:             "primaryip-protection",
			Type:             "ipv4",
			Labels:           nil,
			Datacenter:       teste2e.TestDataCenter,
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
		Datacenter:       teste2e.TestDataCenter,
		AssigneeType:     "server",
		DeleteProtection: false,
	}

	protected := &primaryip.RData{
		Name:             unprotected.Name,
		Type:             unprotected.Type,
		Datacenter:       unprotected.Datacenter,
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
