package loadbalancertype_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
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
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_load_balancer_type", byName,
					"testdata/d/hcloud_load_balancer_type", byID,
				),

				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("id"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("name"), knownvalue.StringExact("lb11")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("LB11")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("max_assigned_certificates"), knownvalue.Int64Exact(10)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("max_connections"), knownvalue.Int64Exact(10000)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("max_services"), knownvalue.Int64Exact(5)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("max_targets"), knownvalue.Int64Exact(25)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("is_deprecated"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("deprecation_announced"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("unavailable_after"), knownvalue.Null()),

					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("id"), knownvalue.Int64Exact(1)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("name"), knownvalue.StringExact("lb11")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("description"), knownvalue.StringExact("LB11")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("max_assigned_certificates"), knownvalue.Int64Exact(10)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("max_connections"), knownvalue.Int64Exact(10000)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("max_services"), knownvalue.Int64Exact(5)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("max_targets"), knownvalue.Int64Exact(25)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("is_deprecated"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("deprecation_announced"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("unavailable_after"), knownvalue.Null()),
				},
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
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
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
