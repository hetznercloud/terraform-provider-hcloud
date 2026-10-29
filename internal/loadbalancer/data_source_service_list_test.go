package loadbalancer_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccLoadBalancerServiceListDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resLoadBalancer := &loadbalancer.RData{
		Name:         "some-load-balancer",
		LocationName: teste2e.TestLocationName,
	}
	resLoadBalancer.SetRName("test_lb")

	res1 := &loadbalancer.RDataService{
		LoadBalancerID:  resLoadBalancer.TFID() + ".id",
		Protocol:        "tcp",
		ListenPort:      22,
		DestinationPort: 2222,
		Proxyprotocol:   true,
	}
	res1.SetRName("test_lb_service_1")

	res2 := &loadbalancer.RDataService{
		LoadBalancerID:  resLoadBalancer.TFID() + ".id",
		Protocol:        "http",
		ListenPort:      80,
		DestinationPort: 8080,
		Proxyprotocol:   false,
	}
	res2.SetRName("test_lb_service_2")

	byLoadBalancerID := &loadbalancer.DDataServiceList{
		LoadBalancerID: resLoadBalancer.TFID() + ".id",
	}
	byLoadBalancerID.SetRName("by_lb_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(loadbalancer.ResourceType, loadbalancer.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_service", res1,
					"testdata/r/hcloud_load_balancer_service", res2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_service", res1,
					"testdata/r/hcloud_load_balancer_service", res2,
					"testdata/d/hcloud_load_balancer_services", byLoadBalancerID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services"), knownvalue.ListSizeExact(2)),

					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(0).AtMapKey("protocol"), knownvalue.StringExact(res1.Protocol)),
					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(0).AtMapKey("listen_port"), knownvalue.Int32Exact(int32(res1.ListenPort))),
					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(0).AtMapKey("destination_port"), knownvalue.Int32Exact(int32(res1.DestinationPort))),
					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(0).AtMapKey("proxyprotocol"), knownvalue.Bool(res1.Proxyprotocol)),

					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(1).AtMapKey("protocol"), knownvalue.StringExact(res2.Protocol)),
					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(1).AtMapKey("listen_port"), knownvalue.Int32Exact(int32(res2.ListenPort))),
					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(1).AtMapKey("destination_port"), knownvalue.Int32Exact(int32(res2.DestinationPort))),
					statecheck.ExpectKnownValue(byLoadBalancerID.TFID(), tfjsonpath.New("services").AtSliceIndex(1).AtMapKey("proxyprotocol"), knownvalue.Bool(res2.Proxyprotocol)),
				},
			},
		},
	})
}
