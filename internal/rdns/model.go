package rdns

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type model struct {
	ID             types.String      `tfsdk:"id"`
	ServerID       types.Int64       `tfsdk:"server_id"`
	PrimaryIPID    types.Int64       `tfsdk:"primary_ip_id"`
	FloatingIPID   types.Int64       `tfsdk:"floating_ip_id"`
	LoadBalancerID types.Int64       `tfsdk:"load_balancer_id"`
	IPAddress      iptypes.IPAddress `tfsdk:"ip_address"`
	DNSPtr         types.String      `tfsdk:"dns_ptr"`
}

func (m *model) FromAPI(_ context.Context, rdns hcloud.RDNSSupporter, ip net.IP, dnsPtr string) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(FormatID(rdns, ip))

	m.ServerID = types.Int64Null()
	m.PrimaryIPID = types.Int64Null()
	m.FloatingIPID = types.Int64Null()
	m.LoadBalancerID = types.Int64Null()

	switch v := rdns.(type) {
	case *hcloud.Server:
		m.ServerID = types.Int64Value(v.ID)
	case *hcloud.PrimaryIP:
		m.PrimaryIPID = types.Int64Value(v.ID)
	case *hcloud.FloatingIP:
		m.FloatingIPID = types.Int64Value(v.ID)
	case *hcloud.LoadBalancer:
		m.LoadBalancerID = types.Int64Value(v.ID)
	}

	m.IPAddress = iptypes.NewIPAddressValue(ip.String())
	m.DNSPtr = types.StringValue(dnsPtr)

	return diags
}
