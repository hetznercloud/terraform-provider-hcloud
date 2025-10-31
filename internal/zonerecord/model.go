package zonerecord

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

type model struct {
	Zone    types.String `tfsdk:"zone"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	TTL     types.Int32  `tfsdk:"ttl"`
	Value   types.String `tfsdk:"value"`
	Comment types.String `tfsdk:"comment"`
}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"zone":    types.StringType,
		"name":    types.StringType,
		"type":    types.StringType,
		"ttl":     types.Int32Type,
		"value":   types.StringType,
		"comment": types.StringType,
	}
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

// var _ util.ModelFromAPI[*hcloud.ZoneRRSet] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) FromAPI(_ context.Context, hc *hcloud.ZoneRRSet) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = types.StringValue(hc.Name)
	m.Type = types.StringValue(string(hc.Type))

	if hc.TTL != nil {
		m.TTL = types.Int32Value(int32(*hc.TTL)) // nolint: gosec
	} else {
		m.TTL = types.Int32Null()
	}

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

func (i *identityModel) ToAPI(_ context.Context) (*hcloud.ZoneRRSet, string) {
	return &hcloud.ZoneRRSet{
		Zone: &hcloud.Zone{Name: i.Zone.ValueString()},
		Name: i.Name.ValueString(),
		Type: hcloud.ZoneRRSetType(i.Type.ValueString()),
	}, i.Value.ValueString()
}

func (i *identityModel) FromAPI(_ context.Context, in *hcloud.ZoneRRSet, value string) {
	i.Zone = types.StringValue(in.Zone.Name)
	i.Name = types.StringValue(in.Name)
	i.Type = types.StringValue(string(in.Type))
	// TODO: TXT transformation
	i.Value = types.StringValue(value)
}
