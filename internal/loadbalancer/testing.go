package loadbalancer

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/certificate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name: ResourceType,
		Dependencies: []string{
			certificate.ResourceType,
			server.ResourceType,
			network.ResourceType,
		},
		F: Sweep,
	})
}

// Basic Load Balancer for use in load balancer related test.
//
// Do not modify!
var Basic = &RData{
	Name:         "basic-load-balancer",
	LocationName: "nbg1",
}

// Sweep removes all Load Balancers from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	loadBalancers, err := client.LoadBalancer.All(ctx)
	if err != nil {
		return err
	}

	for _, loadBalancer := range loadBalancers {
		if _, err := client.LoadBalancer.Delete(ctx, loadBalancer); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a loadbalancer by its ID.
func ByID(t *testing.T, lb *hcloud.LoadBalancer) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.LoadBalancer.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find load balancer %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if lb != nil {
			*lb = *found
		}
		return true
	}
}

// DData defines the fields for the "testdata/d/hcloud_load_balancer"
// template.
type DData struct {
	testtemplate.DataCommon

	Name             string
	LoadBalancerID   string
	LoadBalancerName string
	LabelSelector    string
}

// TFID returns the data source identifier.
func (d *DData) TFID() string {
	return fmt.Sprintf("data.%s.%s", DataSourceType, d.Name)
}

// RData defines the fields for the "testdata/r/hcloud_load_balancer"
// template.
type RData struct {
	testtemplate.DataCommon

	Name          string
	Type          string
	LocationName  string
	NetworkZone   string
	Algorithm     string
	ServerTargets []RDataInlineServerTarget
	Labels        map[string]string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.Name)
}

// RDataInlineServerTarget represents a Load Balancer server target
// that is added inline to the Load Balancer.
type RDataInlineServerTarget struct {
	ServerID string
}

// RDataService defines the fields for the
// "testdata/r/hcloud_load_balancer_service" template.
type RDataService struct {
	testtemplate.DataCommon

	Name            string
	LoadBalancerID  string
	Protocol        string
	ListenPort      int
	DestinationPort int
	Proxyprotocol   bool

	AddHTTP bool // Required as the RLoadBalancerServiceHTTP is not comparable
	HTTP    RDataServiceHTTP

	AddHealthCheck bool // Required as the RLoadBalancerServiceHealthCheck is not comparable
	HealthCheck    RDataServiceHealthCheck
}

// RDataServiceHTTP contains data for an HTTP load balancer service.
type RDataServiceHTTP struct {
	CookieName     string
	CookieLifeTime int
	Certificates   []string
	RedirectHTTP   bool
	StickySessions bool
}

// RDataServiceHealthCheck contains data for a load balancer service
// Health Check.
type RDataServiceHealthCheck struct {
	Protocol string
	Port     int
	Interval int
	Timeout  int
	Retries  int
	HTTP     RDataServiceHealthCheckHTTP
}

// RDataServiceHealthCheckHTTP contains data for a load balancer service
// HTTP Health Check.
type RDataServiceHealthCheckHTTP struct {
	Domain      string
	Path        string
	Response    string
	TLS         bool
	StatusCodes []string
}

// RDataTarget defines the fields for the
// "testdata/r/hcloud_load_balancer_target" template.
type RDataTarget struct {
	testtemplate.DataCommon

	Name           string
	Type           string
	LoadBalancerID string
	ServerID       string
	LabelSelector  string
	IP             string
	UsePrivateIP   bool
	DependsOn      []string
}

// RDataNetwork defines the fields for the
// "testdata/r/hcloud_load_balancer_network" template.
type RDataNetwork struct {
	testtemplate.DataCommon

	Name                  string
	LoadBalancerID        string
	NetworkID             string
	SubNetID              string
	IP                    string
	EnablePublicInterface bool
	DependsOn             []string
}
