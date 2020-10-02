package loadbalancer_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudLoadBalancerService_TCP(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadbalancer.Basic.Name)
	svcName := "lb-tcp-service-test"
	svcResName := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName,
						Protocol:        "tcp",
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						ListenPort:      70,
						DestinationPort: 70,
						Proxyprotocol:   true,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 70)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "70"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "70"),
					resource.TestCheckResourceAttr(svcResName, "proxyprotocol", "true"),
				),
			},
			{
				// Try to import the newly created volume attachment
				ResourceName:      svcResName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return fmt.Sprintf("%d__%d", lb.ID, 70), nil
				},
			},
			{ // Test disable Proxyprotocol
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName,
						Protocol:        "tcp",
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						ListenPort:      70,
						DestinationPort: 70,
						Proxyprotocol:   false,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 70)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "70"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "70"),
					resource.TestCheckResourceAttr(svcResName, "proxyprotocol", "false"),
				),
			},
		},
	})
}

func TestAccHcloudLoadBalancerService_HTTP(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadbalancer.Basic.Name)
	svcName := "lb-http-service-test"
	svcResName := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a HTTP service using defaults
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						Protocol:       "http",
						LoadBalancerID: fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 80)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "http"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "80"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "80"),
				),
			},
			{
				// Create a HTTP service using non-default ports.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName,
						Protocol:        "http",
						ListenPort:      81,
						DestinationPort: 8080,
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						AddHTTP:         true,
						HTTP: loadbalancer.RDataServiceHTTP{
							CookieName:     "TESTCOOKIE",
							CookieLifeTime: 800,
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 81)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "http"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "81"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "8080"),
					resource.TestCheckResourceAttr(svcResName, "http.0.cookie_name", "TESTCOOKIE"),
					resource.TestCheckResourceAttr(svcResName, "http.0.cookie_lifetime", "800"),
				),
			},
			{
				// Create a HTTP service with health check
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName,
						Protocol:        "http",
						DestinationPort: 8080,
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						AddHealthCheck:  true,
						HealthCheck: loadbalancer.RDataServiceHealthCheck{
							Protocol: "http",
							Port:     8080,
							Interval: 30,
							Timeout:  20,
							Retries:  2,
							HTTP: loadbalancer.RDataServiceHealthCheckHTTP{
								Domain:      "example.com",
								Path:        "/internal/health",
								Response:    "OK",
								StatusCodes: []string{"2??", "301"},
							},
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 81)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "http"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "81"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "8080"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.protocol", "http"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.port", "8080"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.timeout", "20"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.retries", "2"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.http.0.domain", "example.com"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.http.0.path", "/internal/health"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.http.0.response", "OK"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.http.0.status_codes.0", "2??"),
					resource.TestCheckResourceAttr(svcResName, "health_check.0.http.0.status_codes.1", "301"),
				),
			},
		},
	})
}

func TestAccHcloudLoadBalancerService_HTTPS(t *testing.T) {
	var (
		lb   hcloud.LoadBalancer
		cert hcloud.Certificate
	)

	certData := certificate.NewRData(t, "test-cert", "example.org")
	certResName := fmt.Sprintf("%s.%s", certificate.ResourceType, certData.RName())

	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadbalancer.Basic.Name)
	svcName := "lb-https-service-test"
	svcResName := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_certificate", certData,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						LoadBalancerID: lbResName + ".id",
						Protocol:       "https",
						AddHTTP:        true,
						HTTP: loadbalancer.RDataServiceHTTP{
							Certificates: []string{certResName + ".id"},
							RedirectHTTP: true,
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(certResName, certificate.ByID(t, &cert)),
					testsupport.LiftTCF(hasService(&lb, 443)),
					testsupport.CheckResourceAttrFunc(svcResName, "http.0.certificates.0", func() string {
						return strconv.Itoa(cert.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "https"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "443"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "80"),
				),
			},
		},
	})
}

func TestAccHcloudLoadBalancerService_CreateDelete_NoListenPort(t *testing.T) {
	svcName := "lb-create-delete-service-test"

	tmplMan := testtemplate.Manager{}

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a HTTP service without setting a listen port.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						Protocol:       "http",
						LoadBalancerID: fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
					},
				),
			},
			{
				// Immediately remove it from the Load Balancer.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", loadbalancer.Basic),
			},
		},
	})

	certData := certificate.NewRData(t, "test-cert", "example.org")
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a HTTPS service without setting a listen port.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_certificate", certData,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						Protocol:       "https",
						LoadBalancerID: fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						AddHTTP:        true,
						HTTP: loadbalancer.RDataServiceHTTP{
							Certificates: []string{fmt.Sprintf("hcloud_certificate.%s.id", certData.Name)},
						},
					},
				),
			},
			{
				// Immediately remove it from the Load Balancer.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", loadbalancer.Basic),
			},
		},
	})
}

func TestAccHcloudLoadBalancerService_ChangeListenPort(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadbalancer.Basic.Name)
	svcName := "lb-change-listen-port-service-test"
	svcName2 := "lb-change-lp-test"
	svcResName := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName)
	svcResName2 := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName2)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName,
						Protocol:        "tcp",
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						ListenPort:      70,
						DestinationPort: 70,
					},
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName2,
						Protocol:        "tcp",
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						ListenPort:      443,
						DestinationPort: 443,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 70)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "70"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "70"),

					testsupport.LiftTCF(hasService(&lb, 443)),
					testsupport.CheckResourceAttrFunc(svcResName2, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName2, "protocol", "tcp"),
					resource.TestCheckResourceAttr(svcResName2, "listen_port", "443"),
					resource.TestCheckResourceAttr(svcResName2, "destination_port", "443"),
				),
			},

			{ // Test Change Listenport
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName,
						Protocol:        "tcp",
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						ListenPort:      71,
						DestinationPort: 70,
					},
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:            svcName2,
						Protocol:        "tcp",
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						ListenPort:      443,
						DestinationPort: 443,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 71)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "71"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "70"),

					testsupport.LiftTCF(hasService(&lb, 443)),
					testsupport.CheckResourceAttrFunc(svcResName2, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName2, "protocol", "tcp"),
					resource.TestCheckResourceAttr(svcResName2, "listen_port", "443"),
					resource.TestCheckResourceAttr(svcResName2, "destination_port", "443"),
				),
			},
		},
	})
}

func hasService(lb *hcloud.LoadBalancer, listenPort int) func() error {
	return func() error {
		for _, svc := range lb.Services {
			if svc.ListenPort == listenPort {
				return nil
			}
		}
		return fmt.Errorf("listen port %d: service not found", listenPort)
	}
}
