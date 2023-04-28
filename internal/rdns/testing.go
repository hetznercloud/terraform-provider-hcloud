package rdns

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name:         ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all sshkeys from the Hetzner Cloud backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	sshkeys, err := client.SSHKey.All(ctx)
	if err != nil {
		return err
	}

	for _, cert := range sshkeys {
		if _, err := client.SSHKey.Delete(ctx, cert); err != nil {
			return err
		}
	}

	return nil
}

// RData defines the fields for the "testdata/r/hcloud_rdns"
// template.
type RData struct {
	testtemplate.DataCommon

	ServerID       string
	PrimaryIPID    string
	FloatingIPID   string
	LoadBalancerID string
	IPAddress      string
	DNSPTR         string
}

// TFID returns the resource identifier.
func (d *RData) TFID() string {
	return fmt.Sprintf("%s.%s", ResourceType, d.RName())
}

// NewRDataServer creates data for a new rdns resource with server_id.
func NewRDataServer(t *testing.T, rName string, serverID string, ipAddress string, dnsPTR string) *RData {
	r := &RData{
		ServerID:  serverID,
		IPAddress: ipAddress,
		DNSPTR:    dnsPTR,
	}
	r.SetRName(rName)
	return r
}

// NewRDataPrimaryIP creates data for a new rdns resource with primary_ip_id.
func NewRDataPrimaryIP(t *testing.T, rName string, primaryIPID string, ipAddress string, dnsPTR string) *RData {
	r := &RData{
		PrimaryIPID: primaryIPID,
		IPAddress:   ipAddress,
		DNSPTR:      dnsPTR,
	}
	r.SetRName(rName)
	return r
}

// NewRDataFloatingIP creates data for a new rdns resource with floating_ip_id.
func NewRDataFloatingIP(t *testing.T, rName string, floatingIPID string, ipAddress string, dnsPTR string) *RData {
	r := &RData{
		FloatingIPID: floatingIPID,
		IPAddress:    ipAddress,
		DNSPTR:       dnsPTR,
	}
	r.SetRName(rName)
	return r
}

// NewRDataLoadBalancer creates data for a new rdns resource with load_balancer_id.
func NewRDataLoadBalancer(t *testing.T, rName string, loadBalancer string, ipAddress string, dnsPTR string) *RData {
	r := &RData{
		LoadBalancerID: loadBalancer,
		IPAddress:      ipAddress,
		DNSPTR:         dnsPTR,
	}
	r.SetRName(rName)
	return r
}
