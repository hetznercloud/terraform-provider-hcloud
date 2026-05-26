package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

func TestAccLoadBalancerServiceResource_TCP(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbRes := LoadBalancerRData()
	lbRes.SetRName("main")

	res1 := &loadbalancer.RDataService{
		Name:            "lb-tcp-service-test",
		Protocol:        "tcp",
		LoadBalancerID:  lbRes.TFID() + ".id",
		ListenPort:      70,
		DestinationPort: 70,
		Proxyprotocol:   true,
	}

	res2 := testtemplate.DeepCopy(t, res1)
	res2.Proxyprotocol = false

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 70)),
					testsupport.CheckResourceAttrFunc(res1.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(res1.TFID(), "protocol", "tcp"),
					resource.TestCheckResourceAttr(res1.TFID(), "listen_port", "70"),
					resource.TestCheckResourceAttr(res1.TFID(), "destination_port", "70"),
					resource.TestCheckResourceAttr(res1.TFID(), "proxyprotocol", "true"),
				),
			},
			{
				// Try to import the newly created load balancer service
				ResourceName:      res1.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%d__%d", lb.ID, 70), nil
				},
			},
			{ // Test disable Proxyprotocol
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 70)),
					testsupport.CheckResourceAttrFunc(res2.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(res2.TFID(), "protocol", "tcp"),
					resource.TestCheckResourceAttr(res2.TFID(), "listen_port", "70"),
					resource.TestCheckResourceAttr(res2.TFID(), "destination_port", "70"),
					resource.TestCheckResourceAttr(res2.TFID(), "proxyprotocol", "false"),
				),
			},
		},
	})
}

func TestAccLoadBalancerServiceResource_HTTP(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbRes := LoadBalancerRData()
	lbRes.SetRName("main")

	res1 := &loadbalancer.RDataService{
		Name:           "lb-http-service-test",
		Protocol:       "http",
		LoadBalancerID: lbRes.TFID() + ".id",
	}

	res2 := testtemplate.DeepCopy(t, res1)
	res2.ListenPort = 81
	res2.DestinationPort = 8080
	res2.AddHTTP = true
	res2.HTTP = loadbalancer.RDataServiceHTTP{
		CookieName:     "TESTCOOKIE",
		CookieLifeTime: 800,
		TimeoutIdle:    60,
	}

	res3 := testtemplate.DeepCopy(t, res1) // Copy from step1
	res3.DestinationPort = 8080
	res3.AddHealthCheck = true
	res3.HealthCheck = loadbalancer.RDataServiceHealthCheck{
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
	}

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a HTTP service using defaults
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 80)),
					testsupport.CheckResourceAttrFunc(res1.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(res1.TFID(), "protocol", "http"),
					resource.TestCheckResourceAttr(res1.TFID(), "listen_port", "80"),
					resource.TestCheckResourceAttr(res1.TFID(), "destination_port", "80"),
				),
			},
			{
				// Create a HTTP service using non-default ports.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 81)),
					testsupport.CheckResourceAttrFunc(res2.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(res2.TFID(), "protocol", "http"),
					resource.TestCheckResourceAttr(res2.TFID(), "listen_port", "81"),
					resource.TestCheckResourceAttr(res2.TFID(), "destination_port", "8080"),
					resource.TestCheckResourceAttr(res2.TFID(), "http.0.cookie_name", "TESTCOOKIE"),
					resource.TestCheckResourceAttr(res2.TFID(), "http.0.cookie_lifetime", "800"),
					resource.TestCheckResourceAttr(res2.TFID(), "http.0.timeout_idle", "60"),
				),
			},
			{
				// Create a HTTP service with health check
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", res3,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 81)),
					testsupport.CheckResourceAttrFunc(res3.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(res3.TFID(), "protocol", "http"),
					resource.TestCheckResourceAttr(res3.TFID(), "listen_port", "81"),
					resource.TestCheckResourceAttr(res3.TFID(), "destination_port", "8080"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.protocol", "http"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.port", "8080"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.timeout", "20"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.retries", "2"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.http.0.domain", "example.com"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.http.0.path", "/internal/health"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.http.0.response", "OK"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.http.0.status_codes.0", "2??"),
					resource.TestCheckResourceAttr(res3.TFID(), "health_check.0.http.0.status_codes.1", "301"),
				),
			},
		},
	})
}

func TestAccLoadBalancerServiceResource_HTTP_StickySessions(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbRes := LoadBalancerRData()
	lbRes.SetRName("main")

	res1 := &loadbalancer.RDataService{
		Name:     "lb-http-sticky-sessions-test",
		Protocol: "http",
		AddHTTP:  true,
		HTTP: loadbalancer.RDataServiceHTTP{
			StickySessions: true,
			CookieLifeTime: 1800,
		},
		LoadBalancerID: lbRes.TFID() + ".id",
	}
	res1.SetRName(res1.Name)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a HTTP service using defaults
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 80)),
					testsupport.CheckResourceAttrFunc(res1.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(res1.TFID(), "protocol", "http"),
					resource.TestCheckResourceAttr(res1.TFID(), "http.0.cookie_lifetime", "1800"),
					resource.TestCheckResourceAttr(res1.TFID(), "http.0.sticky_sessions", "true"),
				),
			},
		},
	})
}

func TestAccLoadBalancerServiceResource_HTTPS(t *testing.T) {
	var (
		lb   hcloud.LoadBalancer
		cert hcloud.Certificate
	)

	certData := certificate.NewUploadedRData(t, "test-cert", "example.org")

	lbRes := LoadBalancerRData()
	lbRes.SetRName("main")

	res1 := &loadbalancer.RDataService{
		Name:           "lb-https-service-test",
		LoadBalancerID: lbRes.TFID() + ".id",
		Protocol:       "https",
		AddHTTP:        true,
		HTTP: loadbalancer.RDataServiceHTTP{
			Certificates: []string{certData.TFID() + ".id"},
			RedirectHTTP: true,
		},
	}
	res1.SetRName(res1.Name)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", certData,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.CheckResourceExists(certData.TFID(), certificate.ByID(t, &cert)),
					testsupport.LiftTCF(hasService(&lb, 443)),
					testsupport.CheckResourceAttrFunc(res1.TFID(), "http.0.certificates.0", func() string {
						return util.FormatID(cert.ID)
					}),
					resource.TestCheckResourceAttr(res1.TFID(), "protocol", "https"),
					resource.TestCheckResourceAttr(res1.TFID(), "listen_port", "443"),
					resource.TestCheckResourceAttr(res1.TFID(), "destination_port", "80"),
				),
			},
		},
	})
}

func TestAccLoadBalancerServiceResource_HTTPS_UpdateUnchangedCertificates(t *testing.T) {
	certRes1 := certificate.NewUploadedRData(t, "cert-res1", "TFAccTests1")
	certRes2 := certificate.NewUploadedRData(t, "cert-res2", "TFAccTests2")
	lbRes := &loadbalancer.RData{
		Name:         "load-balancer-certificates-unchanged",
		LocationName: teste2e.TestLocationName,
	}
	lbRes.SetRName("main")

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
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
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

func TestAccLoadBalancerServiceResource_CreateDelete_NoListenPort(t *testing.T) {
	svcName := "lb-create-delete-service-test"

	certData := certificate.NewUploadedRData(t, "test-cert", "example.org")
	lbRes := LoadBalancerRData()
	lbRes.SetRName("main")

	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			// HTTP
			{
				// Create a HTTP service without setting a listen port.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						Protocol:       "http",
						LoadBalancerID: lbRes.TFID() + ".id",
					},
				),
			},
			{
				// Immediately remove it from the Load Balancer.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", lbRes),
			},

			// HTTPS
			{
				// Create a HTTPS service without setting a listen port.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_uploaded_certificate", certData,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", &loadbalancer.RDataService{
						Name:           svcName,
						Protocol:       "https",
						LoadBalancerID: lbRes.TFID() + ".id",
						AddHTTP:        true,
						HTTP: loadbalancer.RDataServiceHTTP{
							Certificates: []string{certData.TFID() + ".id"},
						},
					},
				),
			},
			{
				// Immediately remove it from the Load Balancer.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", lbRes),
			},
		},
	})
}

func TestAccLoadBalancerServiceResource_ChangeListenPort(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbRes := LoadBalancerRData()
	lbRes.SetRName("main")

	resA1 := &loadbalancer.RDataService{
		Name:            "lb-change-listen-port-service-test",
		Protocol:        "tcp",
		LoadBalancerID:  lbRes.TFID() + ".id",
		ListenPort:      70,
		DestinationPort: 70,
	}
	resA1.SetRName(resA1.Name)

	resA2 := testtemplate.DeepCopy(t, resA1)
	resA2.ListenPort = 71

	resB1 := &loadbalancer.RDataService{
		Name:            "lb-change-lp-test",
		Protocol:        "tcp",
		LoadBalancerID:  lbRes.TFID() + ".id",
		ListenPort:      443,
		DestinationPort: 443,
	}
	resB1.SetRName(resB1.Name)

	resB2 := testtemplate.DeepCopy(t, resB1)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", resA1,
					"testdata/r/hcloud_load_balancer_service", resB1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 70)),
					testsupport.CheckResourceAttrFunc(resA1.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(resA1.TFID(), "protocol", "tcp"),
					resource.TestCheckResourceAttr(resA1.TFID(), "listen_port", "70"),
					resource.TestCheckResourceAttr(resA1.TFID(), "destination_port", "70"),

					testsupport.LiftTCF(hasService(&lb, 443)),
					testsupport.CheckResourceAttrFunc(resB1.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(resB1.TFID(), "protocol", "tcp"),
					resource.TestCheckResourceAttr(resB1.TFID(), "listen_port", "443"),
					resource.TestCheckResourceAttr(resB1.TFID(), "destination_port", "443"),
				),
			},

			{ // Test Change Listenport
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbRes,
					"testdata/r/hcloud_load_balancer_service", resA2,
					"testdata/r/hcloud_load_balancer_service", resB2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbRes.TFID(), loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 71)),
					testsupport.CheckResourceAttrFunc(resA2.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(resA2.TFID(), "protocol", "tcp"),
					resource.TestCheckResourceAttr(resA2.TFID(), "listen_port", "71"),
					resource.TestCheckResourceAttr(resA2.TFID(), "destination_port", "70"),

					testsupport.LiftTCF(hasService(&lb, 443)),
					testsupport.CheckResourceAttrFunc(resB2.TFID(), "load_balancer_id", func() string {
						return util.FormatID(lb.ID)
					}),
					resource.TestCheckResourceAttr(resB2.TFID(), "protocol", "tcp"),
					resource.TestCheckResourceAttr(resB2.TFID(), "listen_port", "443"),
					resource.TestCheckResourceAttr(resB2.TFID(), "destination_port", "443"),
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
