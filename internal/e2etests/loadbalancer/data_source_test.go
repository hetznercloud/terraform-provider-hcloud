package loadbalancer_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceLoadBalancerTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &loadbalancer.RData{
		Name:         "some-load-balancer",
		LocationName: e2etests.TestLocationName,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	lbByName := &loadbalancer.DData{
		LoadBalancerName: res.TFID() + ".name",
	}
	lbByName.SetRName("lb_by_name")
	lbByID := &loadbalancer.DData{
		LoadBalancerID: res.TFID() + ".id",
	}
	lbByID.SetRName("lb_by_id")
	lbBySel := &loadbalancer.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	lbBySel.SetRName("lb_by_sel")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", res,
					"testdata/d/hcloud_load_balancer", lbByName,
					"testdata/d/hcloud_load_balancer", lbByID,
					"testdata/d/hcloud_load_balancer", lbBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lbByName.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(lbByName.TFID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(lbByName.TFID(), "target.#", "0"),
					resource.TestCheckResourceAttr(lbByName.TFID(), "service.#", "0"),

					resource.TestCheckResourceAttr(lbByID.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(lbByID.TFID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(lbByID.TFID(), "target.#", "0"),
					resource.TestCheckResourceAttr(lbByID.TFID(), "service.#", "0"),

					resource.TestCheckResourceAttr(lbBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(lbBySel.TFID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(lbBySel.TFID(), "target.#", "0"),
					resource.TestCheckResourceAttr(lbBySel.TFID(), "service.#", "0"),
				),
			},
		},
	})
}

func TestAccHcloudDataSourceLoadBalancerListTest(t *testing.T) {
	res := &loadbalancer.RData{
		Name:         "some-load-balancer",
		LocationName: e2etests.TestLocationName,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}

	loadBalancersBySel := &loadbalancer.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	loadBalancersBySel.SetRName("load_balancers_by_sel")

	allLoadBalancersSel := &loadbalancer.DDataList{}
	allLoadBalancersSel.SetRName("all_load_balancers_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", res,
					"testdata/d/hcloud_load_balancers", loadBalancersBySel,
					"testdata/d/hcloud_load_balancers", allLoadBalancersSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(loadBalancersBySel.TFID(), "load_balancers.*",
						map[string]string{
							"name":      fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"location":  res.LocationName,
							"target.#":  "0",
							"service.#": "0",
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(allLoadBalancersSel.TFID(), "load_balancers.*",
						map[string]string{
							"name":      fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"location":  res.LocationName,
							"target.#":  "0",
							"service.#": "0",
						},
					),
				),
			},
		},
	})
}
