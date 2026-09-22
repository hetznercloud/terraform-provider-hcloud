package networkmember

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/sliceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

type model struct {
	ID       types.Int64        `tfsdk:"id"`
	Type     types.String       `tfsdk:"type"`
	Subnet   cidrtypes.IPPrefix `tfsdk:"subnet"`
	IP       iptypes.IPAddress  `tfsdk:"ip"`
	AliasIPs types.Set          `tfsdk:"alias_ips"`
	Status   types.String       `tfsdk:"status"`
}

var _ util.ModelFromAPI[*hcloud.NetworkMember] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.Int64Type,
		"type":      types.StringType,
		"subnet":    types.StringType,
		"ip":        types.StringType,
		"alias_ips": types.SetType{ElemType: iptypes.IPAddressType{}},
		"status":    types.StringType,
	}
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.NetworkMember) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Type = types.StringValue(string(hc.Type))
	m.Subnet = cidrtypes.NewIPPrefixValue(hc.Subnet.String())
	m.IP = iptypes.NewIPAddressValue(hc.IP.String())

	{
		elements := sliceutil.Transform(
			hc.AliasIPs,
			func(o net.IP) iptypes.IPAddress { return iptypes.NewIPAddressValue(o.String()) },
		)
		m.AliasIPs, newDiags = types.SetValueFrom(ctx, iptypes.IPAddressType{}, elements)
		diags.Append(newDiags...)
	}

	m.Status = types.StringValue(string(hc.Status))

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
