package server

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/sliceutil"
)

type networkResourceData struct {
	ID         types.String      `tfsdk:"id"`
	ServerID   types.Int64       `tfsdk:"server_id"`
	NetworkID  types.Int64       `tfsdk:"network_id"`
	SubnetID   types.String      `tfsdk:"subnet_id"`
	IP         iptypes.IPAddress `tfsdk:"ip"`
	AliasIPs   types.Set         `tfsdk:"alias_ips"`
	MACAddress types.String      `tfsdk:"mac_address"`
}

func populateNetworkResourceData(
	ctx context.Context,
	data *networkResourceData,
	server *hcloud.Server,
	attachment *hcloud.ServerPrivateNet,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	data.ID = types.StringValue(fmt.Sprintf("%d-%d", server.ID, attachment.Network.ID))
	data.ServerID = types.Int64Value(server.ID)
	data.NetworkID = types.Int64Value(attachment.Network.ID)
	data.IP = iptypes.NewIPAddressValue(attachment.IP.String())
	data.MACAddress = types.StringValue(attachment.MACAddress)

	{
		elements := sliceutil.Transform(
			attachment.Aliases,
			func(o net.IP) iptypes.IPAddress { return iptypes.NewIPAddressValue(o.String()) },
		)
		data.AliasIPs, newDiags = types.SetValueFrom(ctx, iptypes.IPAddressType{}, elements)
		diags.Append(newDiags...)
	}

	return diags
}
