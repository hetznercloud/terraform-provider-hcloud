package loadbalancertype_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccLoadBalancerTypeDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	byName := &loadbalancertype.DData{LoadBalancerTypeName: teste2e.TestLoadBalancerType}
	byName.SetRName("by_name")

	byID := &loadbalancertype.DData{LoadBalancerTypeID: "1"}
	byID.SetRName("by_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_load_balancer_type", byName,
					"testdata/d/hcloud_load_balancer_type", byID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(byName.TFID(), "id", "1"),
					resource.TestCheckResourceAttr(byName.TFID(), "name", "lb11"),
					resource.TestCheckResourceAttr(byName.TFID(), "description", "LB11"),
					resource.TestCheckResourceAttr(byName.TFID(), "max_assigned_certificates", "10"),
					resource.TestCheckResourceAttr(byName.TFID(), "max_connections", "10000"),
					resource.TestCheckResourceAttr(byName.TFID(), "max_services", "5"),
					resource.TestCheckResourceAttr(byName.TFID(), "max_targets", "25"),

					resource.TestCheckResourceAttr(byID.TFID(), "id", "1"),
					resource.TestCheckResourceAttr(byID.TFID(), "name", "lb11"),
					resource.TestCheckResourceAttr(byID.TFID(), "description", "LB11"),
					resource.TestCheckResourceAttr(byID.TFID(), "max_assigned_certificates", "10"),
					resource.TestCheckResourceAttr(byID.TFID(), "max_connections", "10000"),
					resource.TestCheckResourceAttr(byID.TFID(), "max_services", "5"),
					resource.TestCheckResourceAttr(byID.TFID(), "max_targets", "25"),
				),
			},
		},
	})
}

func TestAccLoadBalancerTypeDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	all := &loadbalancertype.DDataList{}
	all.SetRName("all")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_load_balancer_types", all,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(all.TFID(), "load_balancer_types.0.id", "1"),
					resource.TestCheckResourceAttr(all.TFID(), "load_balancer_types.0.name", "lb11"),
					resource.TestCheckResourceAttr(all.TFID(), "load_balancer_types.0.description", "LB11"),
					resource.TestCheckResourceAttr(all.TFID(), "load_balancer_types.0.max_assigned_certificates", "10"),
					resource.TestCheckResourceAttr(all.TFID(), "load_balancer_types.0.max_connections", "10000"),
					resource.TestCheckResourceAttr(all.TFID(), "load_balancer_types.0.max_services", "5"),
					resource.TestCheckResourceAttr(all.TFID(), "load_balancer_types.0.max_targets", "25"),
				),
			},
		},
	})
}
