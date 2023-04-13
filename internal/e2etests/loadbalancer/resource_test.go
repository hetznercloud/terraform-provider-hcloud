package loadbalancer_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestLoadBalancerResource_Basic(t *testing.T) {
	var lb hcloud.LoadBalancer

	res := loadbalancer.Basic
	resRenamed := &loadbalancer.RData{
		Name:         res.Name + "-renamed",
		LocationName: e2etests.TestLocationName,
		Algorithm:    "least_connections",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, &lb)),
		Steps: []resource.TestStep{
			{
				// Create a new Load Balancer using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), loadbalancer.ByID(t, &lb)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-load-balancer--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "load_balancer_type", e2etests.TestLoadBalancerType),
					resource.TestCheckResourceAttr(res.TFID(), "location", res.LocationName),
				),
			},
			{
				// Try to import the newly created load balancer
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the Load Balancer created in the previous step by
				// setting all optional fields and renaming the load
				// balancer.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("basic-load-balancer-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "load_balancer_type", e2etests.TestLoadBalancerType),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "location", resRenamed.LocationName),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "algorithm.0.type", "least_connections"),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "labels.key2", "value2"),
				),
			},
		},
	})
}

func TestLoadBalancerResource_Resize(t *testing.T) {
	var lb hcloud.LoadBalancer

	res := loadbalancer.Basic
	resResized := &loadbalancer.RData{
		Name:         res.Name,
		LocationName: e2etests.TestLocationName,
		Type:         "lb21",
	}

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, &lb)),
		Steps: []resource.TestStep{
			{
				// Create a new Load Balancer using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), loadbalancer.ByID(t, &lb)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-load-balancer--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "load_balancer_type", e2etests.TestLoadBalancerType),
					resource.TestCheckResourceAttr(res.TFID(), "location", res.LocationName),
				),
			},
			{
				// Update the Load Balancer created in the previous step by
				// setting another load balancer type.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resResized,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resResized.TFID(), "name",
						fmt.Sprintf("basic-load-balancer--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resResized.TFID(), "load_balancer_type", "lb21"),
					resource.TestCheckResourceAttr(resResized.TFID(), "location", resResized.LocationName),
				),
			},
		},
	})
}

func TestLoadBalancerResource_InlineTarget(t *testing.T) {
	var srv0, srv1 hcloud.Server

	tmplMan := testtemplate.Manager{RandInt: acctest.RandInt()}
	resServer1 := &server.RData{
		Name:  "some-server",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
	}
	resServer1.SetRName("some-server")
	resServer2 := &server.RData{
		Name:  "another-server",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
	}
	resServer2.SetRName("another-server")
	res := &loadbalancer.RData{
		Name:         "some-lb",
		LocationName: e2etests.TestLocationName,
		Algorithm:    "least_connections",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		ServerTargets: []loadbalancer.RDataInlineServerTarget{
			{ServerID: resServer1.TFID() + ".id"},
			{ServerID: resServer2.TFID() + ".id"},
		},
	}
	resWithoutTargets := &loadbalancer.RData{
		Name:         "some-lb",
		LocationName: e2etests.TestLocationName,
		Algorithm:    "least_connections",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	resWithoutTargets.SetRName(res.RName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  e2etests.PreCheck(t),
		Providers: e2etests.Providers(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
			testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		),
		Steps: []resource.TestStep{
			{
				// Add two inline targets to the load balancer
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resServer1,
					"testdata/r/hcloud_server", resServer2,
					"testdata/r/hcloud_load_balancer", res,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testsupport.CheckResourceExists(resServer1.TFID(), server.ByID(t, &srv0)),
					testsupport.CheckResourceExists(resServer2.TFID(), server.ByID(t, &srv1)),
					testsupport.CheckResourceAttrFunc(res.TFID(), "target.0.server_id", func() []string {
						return []string{strconv.Itoa(srv0.ID), strconv.Itoa(srv1.ID)}
					}),
					testsupport.CheckResourceAttrFunc(res.TFID(), "target.1.server_id", func() []string {
						return []string{strconv.Itoa(srv0.ID), strconv.Itoa(srv1.ID)}
					}),
				),
			},
			{
				// Remove the targets from the load balancer
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resWithoutTargets,
				),
				Check: resource.TestCheckNoResourceAttr(res.TFID(), "target.%"),
			},
		},
	})
}

func TestLoadBalancerResource_Protection(t *testing.T) {
	var (
		lb hcloud.LoadBalancer

		res = &loadbalancer.RData{
			Name:             "load-balancer-protection",
			LocationName:     "nbg1",
			DeleteProtection: true,
		}

		updateProtection = func(d *loadbalancer.RData, protection bool) *loadbalancer.RData {
			d.DeleteProtection = protection
			return d
		}
	)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, &lb)),
		Steps: []resource.TestStep{
			{
				// Create a new Load Balancer using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), loadbalancer.ByID(t, &lb)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("load-balancer-protection--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "load_balancer_type", e2etests.TestLoadBalancerType),
					resource.TestCheckResourceAttr(res.TFID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
				),
			},
			{
				// Update delete protection
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", updateProtection(res, false),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res.TFID(), "delete_protection", fmt.Sprintf("%t", res.DeleteProtection)),
				),
			},
		},
	})
}
