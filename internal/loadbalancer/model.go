package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type networkResourceData struct {
	ID             types.String      `tfsdk:"id"`
	LoadBalancerID types.Int64       `tfsdk:"load_balancer_id"`
	NetworkID      types.Int64       `tfsdk:"network_id"`
	SubnetID       types.String      `tfsdk:"subnet_id"`
	IP             iptypes.IPAddress `tfsdk:"ip"`

	EnablePublicInterface types.Bool `tfsdk:"enable_public_interface"`
}

// nolint:unparam
func populateNetworkResourceData(
	_ context.Context,
	data *networkResourceData,
	loadBalancer *hcloud.LoadBalancer,
	attachment *hcloud.LoadBalancerPrivateNet,
) diag.Diagnostics {
	data.ID = types.StringValue(fmt.Sprintf("%d-%d", loadBalancer.ID, attachment.Network.ID))
	data.LoadBalancerID = types.Int64Value(loadBalancer.ID)
	data.NetworkID = types.Int64Value(attachment.Network.ID)
	data.IP = iptypes.NewIPAddressValue(attachment.IP.String())

	data.EnablePublicInterface = types.BoolValue(loadBalancer.PublicNet.Enabled)

	return nil
}
