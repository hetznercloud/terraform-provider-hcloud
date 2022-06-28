package primaryip_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

const (
	testDatacenter = "fsn1-dc14"
)

func TestPrimaryIPResource_Basic(t *testing.T) {
	var pip hcloud.PrimaryIP

	res := &primaryip.RData{
		Name:         "primaryip-test",
		Type:         "ipv4",
		Labels:       nil,
		Datacenter:   testDatacenter,
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
	resource.Test(t, resource.TestCase{
		PreCheck: e2etests.PreCheck(t),
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"hcloud": func() (*schema.Provider, error) {
				return tfhcloud.Provider(), nil
			},
		},
		CheckDestroy: testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &pip)),
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

func TestPrimaryIPResource_with_server(t *testing.T) {
	var srv hcloud.Server
	var pip hcloud.PrimaryIP
	var pip2 hcloud.PrimaryIP
	primaryIPOneRes := &primaryip.RData{
		Name:         "primaryip-test",
		Type:         "ipv4",
		Labels:       nil,
		Datacenter:   testDatacenter,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPOneRes.SetRName("primary_ip_test")

	primaryIPTwoRes := &primaryip.RData{
		Name:         "primaryip-test_2",
		Type:         "ipv4",
		Labels:       nil,
		Datacenter:   testDatacenter,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPTwoRes.SetRName("primary_ip_test_2")

	testServerRes := &server.RData{
		Name:       "server-test",
		Type:       e2etests.TestServerType,
		Image:      e2etests.TestImage,
		Datacenter: testDatacenter,
		Labels:     nil,
		PublicNet: map[string]string{
			"ipv4": primaryIPOneRes.TFID() + ".id",
		},
	}

	testServerUpdatedRes := &server.RData{
		Name:       testServerRes.Name,
		Type:       testServerRes.Type,
		Image:      testServerRes.Image,
		Datacenter: testServerRes.Datacenter,
		Labels:     testServerRes.Labels,
		PublicNet: map[string]string{
			"ipv4": primaryIPTwoRes.TFID() + ".id",
		},
	}
	testServerUpdatedRes.SetRName(testServerRes.RName())

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck: e2etests.PreCheck(t),
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"hcloud": func() (*schema.Provider, error) {
				return tfhcloud.Provider(), nil
			},
		},
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &pip)),
			testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &pip2)),
		),
		Steps: []resource.TestStep{
			{
				// Create a new primary ip & server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", primaryIPOneRes,
					"testdata/r/hcloud_primary_ip", primaryIPTwoRes,
					"testdata/r/hcloud_server", testServerRes),

				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(primaryIPOneRes.TFID(), primaryip.ByID(t, &pip)),
					testsupport.CheckResourceExists(primaryIPTwoRes.TFID(), primaryip.ByID(t, &pip2)),
					testsupport.CheckResourceExists(testServerRes.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(primaryIPOneRes.TFID(), "name",
						fmt.Sprintf("primaryip-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(testServerRes.TFID(), "name",
						fmt.Sprintf("server-test--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(primaryIPOneRes.TFID(), "type", primaryIPOneRes.Type),
					resource.TestCheckResourceAttr(testServerRes.TFID(), "server_type", testServerRes.Type),
					resource.TestCheckResourceAttr(primaryIPOneRes.TFID(), "assignee_id", strconv.Itoa(pip.ID)),
				),
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_primary_ip", primaryIPOneRes,
					"testdata/r/hcloud_primary_ip", primaryIPTwoRes,
					"testdata/r/hcloud_server", testServerUpdatedRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(primaryIPTwoRes.TFID(), "assignee_id", strconv.Itoa(pip2.ID))),
			},
		},
	})
}

func TestPrimaryIPResource_Protection(t *testing.T) {
	var (
		pip hcloud.PrimaryIP

		res = &primaryip.RData{
			Name:             "primaryip-protection",
			Type:             "ipv4",
			Labels:           nil,
			Datacenter:       testDatacenter,
			AssigneeType:     "server",
			DeleteProtection: true,
		}

		updateProtection = func(d *primaryip.RData, protection bool) *primaryip.RData {
			d.DeleteProtection = protection
			return d
		}
	)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck: e2etests.PreCheck(t),
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"hcloud": func() (*schema.Provider, error) {
				return tfhcloud.Provider(), nil
			},
		},
		CheckDestroy: testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, &pip)),
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
					"testdata/r/hcloud_primary_ip", updateProtection(res, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
				),
			},
		},
	})
}
