package zonerecord

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type model struct {
	Zone    types.String `tfsdk:"zone"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Value   types.String `tfsdk:"value"`
	Comment types.String `tfsdk:"comment"`
}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"zone":    types.StringType,
		"name":    types.StringType,
		"type":    types.StringType,
		"value":   types.StringType,
		"comment": types.StringType,
	}
}

func (m *model) FromAPI(_ context.Context, hc *hcloud.ZoneRRSet) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = types.StringValue(hc.Name)
	m.Type = types.StringValue(string(hc.Type))
	m.Value = types.StringValue(hc.Records[0].Value)
	m.Comment = types.StringValue(hc.Records[0].Comment)

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func (m *model) FromIdentity(_ context.Context, identity identityModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Zone = identity.Zone
	m.Name = identity.Name
	m.Type = identity.Type
	m.Value = identity.Value

	return diags
}

type identityModel struct {
	Zone  types.String `tfsdk:"zone"`
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func newIdentity(m model) identityModel {
	i := identityModel{}
	i.FromModel(m)
	return i
}

func (i *identityModel) FromModel(m model) {
	i.Zone = m.Zone
	i.Name = m.Name
	i.Type = m.Type
	i.Value = m.Value
}

func (i *identityModel) ToAPI(_ context.Context) (*hcloud.ZoneRRSet, string) {
	return &hcloud.ZoneRRSet{
		Zone: &hcloud.Zone{Name: i.Zone.ValueString()},
		Name: i.Name.ValueString(),
		Type: hcloud.ZoneRRSetType(i.Type.ValueString()),
	}, i.Value.ValueString()
}

func (i *identityModel) FromAPI(in *hcloud.ZoneRRSet) {
	i.Name = types.StringValue(in.Name)
	i.Type = types.StringValue(string(in.Type))
	i.Value = types.StringValue(in.Records[0].Value)
}
