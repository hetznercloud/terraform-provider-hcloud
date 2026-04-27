package primaryip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

type model struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	IPAddress        types.String `tfsdk:"ip_address"`
	IPNetwork        types.String `tfsdk:"ip_network"`
	Location         types.String `tfsdk:"location"`
	Datacenter       types.String `tfsdk:"datacenter"`
	AssigneeID       types.Int64  `tfsdk:"assignee_id"`
	AssigneeType     types.String `tfsdk:"assignee_type"`
	AutoDelete       types.Bool   `tfsdk:"auto_delete"`
	Labels           types.Map    `tfsdk:"labels"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
}

var _ util.ModelFromAPI[*hcloud.PrimaryIP] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.Int64Type,
		"name":              types.StringType,
		"type":              types.StringType,
		"ip_address":        types.StringType,
		"ip_network":        types.StringType,
		"location":          types.StringType,
		"datacenter":        types.StringType,
		"assignee_id":       types.Int64Type,
		"assignee_type":     types.StringType,
		"labels":            types.MapType{ElemType: types.StringType},
		"delete_protection": types.BoolType,
		"auto_delete":       types.BoolType,
	}
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.PrimaryIP) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Name = types.StringValue(hc.Name)
	m.Type = types.StringValue(string(hc.Type))

	m.IPAddress = types.StringValue(hc.IP.String())
	if hc.Type == hcloud.PrimaryIPTypeIPv6 {
		m.IPNetwork = types.StringValue(hc.Network.String())
	} else {
		m.IPNetwork = types.StringNull()
	}

	m.Location = types.StringValue(hc.Location.Name)
	// Pass through datacenter name as long as it is returned from the API.
	//
	// If the attribute is not returned from the API, we never set the attribute,
	// so whatever is in the state or user config is kept.
	//
	// See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters
	//nolint:staticcheck // Backwards-compatibility
	if hc.Datacenter != nil {
		//nolint:staticcheck // Backwards-compatibility
		m.Datacenter = types.StringValue(hc.Datacenter.Name)
	} else {
		m.Datacenter = types.StringNull()
	}

	m.AssigneeID = types.Int64Value(hc.AssigneeID)
	m.AssigneeType = types.StringValue(hc.AssigneeType)

	m.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, hc.Labels)
	diags.Append(newDiags...)

	m.DeleteProtection = types.BoolValue(hc.Protection.Delete)
	m.AutoDelete = types.BoolValue(hc.AutoDelete)

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
