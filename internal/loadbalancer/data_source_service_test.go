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

func TestAccLoadBalancerServiceDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resLoadBalancer := &loadbalancer.RData{
		Name:         "some-load-balancer",
		LocationName: teste2e.TestLocationName,
	}
	resLoadBalancer.SetRName("test_lb")

	res := &loadbalancer.RDataService{
		LoadBalancerID:  resLoadBalancer.TFID() + ".id",
		Protocol:        "http",
		ListenPort:      80,
		DestinationPort: 8000,
		Proxyprotocol:   false,
		AddHTTP:         true,
		HTTP: loadbalancer.RDataServiceHTTP{
			CookieName:     "HCLBSTICKY",
			CookieLifeTime: 300,
			RedirectHTTP:   false,
			StickySessions: true,
			TimeoutIdle:    50,
		},
		AddHealthCheck: true,
		HealthCheck: loadbalancer.RDataServiceHealthCheck{
			Protocol: "http",
			Port:     4711,
			Interval: 15,
			Timeout:  10,
			Retries:  3,
			HTTP: loadbalancer.RDataServiceHealthCheckHTTP{
				Domain:      "example.com",
				Path:        "/",
				Response:    `ok`,
				StatusCodes: []string{"2??", "3??"},
				TLS:         false,
			},
		},
	}
	res.SetRName("test_lb_service")

	byListenPort := &loadbalancer.DDataService{
		LoadBalancerID: resLoadBalancer.TFID() + ".id",
		ListenPort:     res.TFID() + ".listen_port",
	}
	byListenPort.SetRName("by_listen_port")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(loadbalancer.ResourceType, loadbalancer.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resLoadBalancer,
					"testdata/r/hcloud_load_balancer_service", res,
					"testdata/d/hcloud_load_balancer_service", byListenPort,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("protocol"), knownvalue.StringExact(res.Protocol)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("listen_port"), knownvalue.Int32Exact(int32(res.ListenPort))),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("destination_port"), knownvalue.Int32Exact(int32(res.DestinationPort))),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("proxyprotocol"), knownvalue.Bool(false)),

					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("http").AtMapKey("cookie_name"), knownvalue.StringExact(res.HTTP.CookieName)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("http").AtMapKey("cookie_lifetime"), knownvalue.Int32Exact(int32(res.HTTP.CookieLifeTime))),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("http").AtMapKey("redirect_http"), knownvalue.Bool(res.HTTP.RedirectHTTP)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("http").AtMapKey("sticky_sessions"), knownvalue.Bool(res.HTTP.StickySessions)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("http").AtMapKey("timeout_idle"), knownvalue.Int32Exact(int32(res.HTTP.TimeoutIdle))),

					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("protocol"), knownvalue.StringExact(res.HealthCheck.Protocol)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("port"), knownvalue.Int32Exact(int32(res.HealthCheck.Port))),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("interval"), knownvalue.Int32Exact(int32(res.HealthCheck.Interval))),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("timeout"), knownvalue.Int32Exact(int32(res.HealthCheck.Timeout))),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("retries"), knownvalue.Int32Exact(int32(res.HealthCheck.Retries))),

					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("http").AtMapKey("domain"), knownvalue.StringExact(res.HealthCheck.HTTP.Domain)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("http").AtMapKey("path"), knownvalue.StringExact(res.HealthCheck.HTTP.Path)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("http").AtMapKey("response"), knownvalue.StringExact(res.HealthCheck.HTTP.Response)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("http").AtMapKey("status_codes"), listExactStrings(res.HealthCheck.HTTP.StatusCodes)),
					statecheck.ExpectKnownValue(byListenPort.TFID(), tfjsonpath.New("health_check").AtMapKey("http").AtMapKey("tls"), knownvalue.Bool(res.HealthCheck.HTTP.TLS)),
				},
			},
		},
	})
}

func listExactStrings(strings []string) knownvalue.Check {
	var checks []knownvalue.Check
	for _, val := range strings {
		checks = append(checks, knownvalue.StringExact(val))
	}
	return knownvalue.ListExact(checks)
}
