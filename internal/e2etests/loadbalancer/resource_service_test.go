package loadbalancer_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
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
				// Try to import the newly created load balancer service
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
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

func TestAccHcloudLoadBalancerService_HTTP_StickySessions(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadbalancer.Basic.Name)
	svcName := "lb-http-sticky-sessions-test"
	svcResName := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a HTTP service using defaults
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:     svcName,
						Protocol: "http",
						AddHTTP:  true,
						HTTP: loadbalancer.RDataServiceHTTP{
							StickySessions: true,
							CookieLifeTime: 1800,
						},
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
					resource.TestCheckResourceAttr(svcResName, "http.0.cookie_lifetime", "1800"),
					resource.TestCheckResourceAttr(svcResName, "http.0.sticky_sessions", "true"),
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

	certData := certificate.NewUploadedRData(t, "test-cert", "example.org")

	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadbalancer.Basic.Name)
	svcName := "lb-https-service-test"
	svcResName := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", certData,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						LoadBalancerID: lbResName + ".id",
						Protocol:       "https",
						AddHTTP:        true,
						HTTP: loadbalancer.RDataServiceHTTP{
							Certificates: []string{certData.TFID() + ".id"},
							RedirectHTTP: true,
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(certData.TFID(), certificate.ByID(t, &cert)),
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

func TestAccHcloudLoadBalancerService_HTTPS_UpdateUnchangedCertificates(t *testing.T) {
	certRes1 := certificate.NewUploadedRData(t, "cert-res1", "TFAccTests1")
	certRes2 := certificate.NewUploadedRData(t, "cert-res2", "TFAccTests2")
	lbRes := &loadbalancer.RData{
		Name:         "load-balancer-certificates-unchanged",
		LocationName: e2etests.TestLocationName,
	}
	svcRes := &loadbalancer.RDataService{
		Name:           "service-with-two-certs",
		LoadBalancerID: lbRes.TFID() + ".id",
		Protocol:       "https",
		AddHTTP:        true,
		HTTP: loadbalancer.RDataServiceHTTP{
			Certificates: []string{certRes1.TFID() + ".id", certRes2.TFID() + ".id"},
			RedirectHTTP: true,
		},
	}

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a new Load Balancer using two certificates
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", certRes1,
					"testdata/r/hcloud_uploaded_certificate", certRes2,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", svcRes,
				),
			},
		},
	})
}

func TestAccHcloudLoadBalancerService_CreateDelete_NoListenPort(t *testing.T) {
	svcName := "lb-create-delete-service-test"

	certData := certificate.NewUploadedRData(t, "test-cert", "example.org")

	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			// HTTP
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

			// HTTPS
			{
				// Create a HTTPS service without setting a listen port.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", certData,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						Protocol:       "https",
						LoadBalancerID: fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						AddHTTP:        true,
						HTTP: loadbalancer.RDataServiceHTTP{
							Certificates: []string{fmt.Sprintf("hcloud_uploaded_certificate.%s.id", certData.Name)},
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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
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
