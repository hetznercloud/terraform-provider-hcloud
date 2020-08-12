package loadbalancer_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceLoadBalancerTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &loadbalancer.RData{
		Name:         "some-load-balancer",
		LocationName: "nbg1",
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	lbByName := &loadbalancer.DData{
		Name:             "lb_by_name",
		LoadBalancerName: res.TFID() + ".name",
	}
	lbByID := &loadbalancer.DData{
		Name:           "lb_by_id",
		LoadBalancerID: res.TFID() + ".id",
	}
	lbBySel := &loadbalancer.DData{
		Name:          "lb_by_sel",
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
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
