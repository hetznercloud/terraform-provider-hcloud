package network

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestNetworkResource(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		var cert hcloud.Network

		res := &RData{
			Name:    "network-test",
			IPRange: "10.0.0.0/8",
			Labels:  nil,
		}
		resRenamed := &RData{Name: res.Name + "-renamed", IPRange: res.IPRange}
		resRenamed.SetRName(res.Name)
		tmplMan := testtemplate.Manager{}
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(network.ResourceType, ByID(t, &cert)),
			Steps: []resource.TestStep{
				{
					// Create a new Network using the required values
					// only.
					Config: tmplMan.Render(t, "testdata/r/hcloud_network", res),
					Check: resource.ComposeTestCheckFunc(
						testsupport.CheckResourceExists(res.TFID(), ByID(t, &cert)),
						resource.TestCheckResourceAttr(res.TFID(), "name",
							fmt.Sprintf("network-test--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(res.TFID(), "ip_range", res.IPRange),
					),
				},
				{
					// Try to import the newly created Network
					ResourceName:      res.TFID(),
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					// Update the Network created in the previous step by
					// setting all optional fields and renaming the Network.
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", resRenamed,
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
							fmt.Sprintf("network-test-renamed--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(resRenamed.TFID(), "ip_range", res.IPRange),
					),
				},
			},
		})
	})

	t.Run("increase-network", func(t *testing.T) {
		var cert hcloud.Network

		res := &RData{
			Name:    "network-test-increase",
			IPRange: "10.0.0.0/16",
			Labels:  nil,
		}
		resResized := &RData{Name: res.Name, IPRange: "10.0.0.0/8"}
		resResized.SetRName(res.RName())
		tmplMan := testtemplate.Manager{}
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(network.ResourceType, ByID(t, &cert)),
			Steps: []resource.TestStep{
				{
					// Create a new Network using the required values
					// only.
					Config: tmplMan.Render(t, "testdata/r/hcloud_network", res),
					Check: resource.ComposeTestCheckFunc(
						testsupport.CheckResourceExists(res.TFID(), ByID(t, &cert)),
						resource.TestCheckResourceAttr(res.TFID(), "name",
							fmt.Sprintf("network-test-increase--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(res.TFID(), "ip_range", res.IPRange),
					),
				},
				{
					// Try to import the newly created Network
					ResourceName:      res.TFID(),
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					// Update the Network created in the previous step by
					// setting all optional fields and renaming the Network.
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", resResized,
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(resResized.TFID(), "name",
							fmt.Sprintf("network-test-increase--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(resResized.TFID(), "ip_range", "10.0.0.0/8"),
					),
				},
			},
		})
	})

	t.Run("protection", func(t *testing.T) {
		var (
			cert hcloud.Network

			res = &RData{
				Name:             "network-protection",
				IPRange:          "10.0.0.0/8",
				Labels:           nil,
				DeleteProtection: true,
			}

			updateProtection = func(d *RData, protection bool) *RData {
				d.DeleteProtection = protection
				return d
			}
		)

		tmplMan := testtemplate.Manager{}
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(network.ResourceType, ByID(t, &cert)),
			Steps: []resource.TestStep{
				{
					// Create a new Network using the required values
					// only.
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", res,
					),
					Check: resource.ComposeTestCheckFunc(
						testsupport.CheckResourceExists(res.TFID(), ByID(t, &cert)),
						resource.TestCheckResourceAttr(res.TFID(), "name",
							fmt.Sprintf("network-protection--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(res.TFID(), "ip_range", res.IPRange),
						resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
					),
				},
				{
					// Update delete protection
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", updateProtection(res, false),
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
					),
				},
			},
		})
	})

	t.Run("expose-route-to-v-switch", func(t *testing.T) {
		var (
			cert hcloud.Network

			res = &RData{
				Name:                  "network-routes-vswitch",
				IPRange:               "10.0.0.0/8",
				ExposeRoutesToVSwitch: true,
			}

			updateExposure = func(d *RData, expose bool) *RData {
				d.ExposeRoutesToVSwitch = expose
				return d
			}
		)

		tmplMan := testtemplate.Manager{}
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(network.ResourceType, ByID(t, &cert)),
			Steps: []resource.TestStep{
				{
					// Create a new Network using the required values only.
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", res,
					),
					Check: resource.ComposeTestCheckFunc(
						testsupport.CheckResourceExists(res.TFID(), ByID(t, &cert)),
						resource.TestCheckResourceAttr(res.TFID(), "name",
							fmt.Sprintf("network-routes-vswitch--%d", tmplMan.RandInt)),
						resource.TestCheckResourceAttr(res.TFID(), "ip_range", res.IPRange),
						resource.TestCheckResourceAttr(res.TFID(), "expose_routes_to_vswitch", fmt.Sprintf("%t", res.ExposeRoutesToVSwitch)),
					),
				},
				{
					// Try to import the newly created Network
					ResourceName:      res.TFID(),
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					// Update expose_routes_to_vswitch
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_network", updateExposure(res, false),
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(res.TFID(), "expose_routes_to_vswitch", fmt.Sprintf("%t", res.ExposeRoutesToVSwitch)),
					),
				},
				{
					// Try to import the newly created Network
					ResourceName:      res.TFID(),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
