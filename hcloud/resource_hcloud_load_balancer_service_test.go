package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAccHcloudLoadBalancerService_Create(t *testing.T) {
	var loadBalancer hcloud.LoadBalancer

	rInt := acctest.RandInt()
	rCert, rKey, err := acctest.RandTLSCert("example.org")
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerService_RequiredOnly(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.test_load_balancer", &loadBalancer),
					testAccHcloudCheckLoadBalancerServiceExists(
						"hcloud_load_balancer_service.test_load_balancer_service", 80),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "protocol", "http"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "listen_port", "80"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "destination_port", "80"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "proxyprotocol", "false"),
				),
			},
			{
				Config: testAccLoadBalancerService_WithOptionalAttrs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.test_load_balancer", &loadBalancer),
					testAccHcloudCheckLoadBalancerServiceExists(
						"hcloud_load_balancer_service.test_load_balancer_service", 81),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "protocol", "http"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "listen_port", "81"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "destination_port", "82"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service", "proxyprotocol", "true"),
				),
			},
			{
				Config: testAccLoadBalancerService_HTTPS(rInt, rKey, rCert),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.test_load_balancer", &loadBalancer),
					testAccHcloudCheckLoadBalancerServiceExists(
						"hcloud_load_balancer_service.test_load_balancer_service_https", 443),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_https", "protocol", "https"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_https", "listen_port", "443"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_https", "http.#", "1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_https", "http.0.cookie_name", "MYCOOKIE"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_https", "http.0.cookie_lifetime", "300"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_https", "http.0.certificates.#", "1"),
				),
			},
			{
				Config: testAccLoadBalancerService_HealthCheck(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckLoadBalancerExists("hcloud_load_balancer.test_load_balancer", &loadBalancer),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_with_health_check", "health_check.#", "1"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_with_health_check", "health_check.0.port", "80"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_with_health_check", "health_check.0.http.0.domain", "example.org"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_with_health_check", "health_check.0.http.0.path", "/internal/health"),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_with_health_check", "health_check.0.http.0.response", `{"status": "ok"}`),
					resource.TestCheckResourceAttr(
						"hcloud_load_balancer_service.test_load_balancer_service_with_health_check", "health_check.0.http.0.status_codes.0", "200"),
				),
			},
		},
	})
}

func testAccLoadBalancerServiceLoadBalancer(rInt int) string {
	return fmt.Sprintf(`
	resource "hcloud_load_balancer" "test_load_balancer" {
		name = "test_load_balancer_%d"
		load_balancer_type = "lb11"
		location   = "nbg1"
		algorithm {
			type = "round_robin"
		}
	}`, rInt)
}

func testAccLoadBalancerServiceCertificate(rInt int, key, cert string) string {
	key = fmt.Sprintf("\n%s\n", strings.TrimSpace(key))
	cert = fmt.Sprintf("\n%s\n", strings.TrimSpace(cert))

	return fmt.Sprintf(`
	resource "hcloud_certificate" "test_certificate" {
		name = "test_certificate-%d"
		private_key =<<EOT%sEOT
		certificate =<<EOT%sEOT
	}`, rInt, key, cert)
}

func testAccLoadBalancerService_RequiredOnly(rInt int) string {
	lbSvc := `
	resource "hcloud_load_balancer_service" "test_load_balancer_service" {
		load_balancer_id = "${hcloud_load_balancer.test_load_balancer.id}"
		protocol = "http"
	}`
	return fmt.Sprintf("%s\n%s", testAccLoadBalancerServiceLoadBalancer(rInt), lbSvc)
}

func testAccLoadBalancerService_WithOptionalAttrs(rInt int) string {
	lbSvc := `
	resource "hcloud_load_balancer_service" "test_load_balancer_service" {
		load_balancer_id = "${hcloud_load_balancer.test_load_balancer.id}"
		protocol = "http"
		listen_port = 81
		destination_port = 82
		proxyprotocol = true
	}`
	return fmt.Sprintf("%s\n%s", testAccLoadBalancerServiceLoadBalancer(rInt), lbSvc)
}

func testAccLoadBalancerService_HTTPS(rInt int, key, cert string) string {
	lbSvc := `
	resource "hcloud_load_balancer_service" "test_load_balancer_service_https" {
		load_balancer_id = "${hcloud_load_balancer.test_load_balancer.id}"
		protocol = "https"
		http {
			cookie_name = "MYCOOKIE"
			cookie_lifetime = 300
			certificates = ["${hcloud_certificate.test_certificate.id}"]
			redirect_http = true
		}
	}`
	return fmt.Sprintf("%s\n%s\n%s",
		testAccLoadBalancerServiceCertificate(rInt, key, cert),
		testAccLoadBalancerServiceLoadBalancer(rInt),
		lbSvc)
}

func testAccLoadBalancerService_HealthCheck(rInt int) string {
	lbSvc := `
	resource "hcloud_load_balancer_service" "test_load_balancer_service_with_health_check" {
		load_balancer_id = "${hcloud_load_balancer.test_load_balancer.id}"
		protocol = "http"
		health_check {
			protocol = "http"
			port = 80
			interval = 10
			timeout = 5
			http {
				domain = "example.org"
				path = "/internal/health"
				response = "{\"status\": \"ok\"}"
				status_codes = [200]
			}
		}
	}`
	return fmt.Sprintf("%s\n%s",
		testAccLoadBalancerServiceLoadBalancer(rInt),
		lbSvc)
}

func testAccHcloudCheckLoadBalancerServiceExists(
	n string, listenPort int,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		if rs.Primary.Attributes["load_balancer_id"] == "" {
			return fmt.Errorf("No load_balancer_id set")
		}
		id, err := strconv.Atoi(rs.Primary.Attributes["load_balancer_id"])
		if err != nil {
			return err
		}
		client := testAccProvider.Meta().(*hcloud.Client)
		lb, _, err := client.LoadBalancer.GetByID(context.Background(), id)
		if err != nil {
			return err
		}
		if lb == nil {
			return fmt.Errorf("load balancer not found: %d", id)
		}
		for _, svc := range lb.Services {
			if svc.ListenPort == listenPort {
				return nil
			}
		}
		return fmt.Errorf("load balancer %d: no service for port %d", id, listenPort)
	}
}
