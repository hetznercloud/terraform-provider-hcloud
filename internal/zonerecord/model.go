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

type modelRecord struct {
	Value   types.String `tfsdk:"value"`
	Comment types.String `tfsdk:"comment"`
}

var _ util.ModelFromAPI[hcloud.ZoneRRSetRecord] = &modelRecord{}
var _ util.ModelFromTerraform[types.Object] = &modelRecord{}
var _ util.ModelToAPI[hcloud.ZoneRRSetRecord] = &modelRecord{}
var _ util.ModelToTerraform[types.Object] = &modelRecord{}

func (m *modelRecord) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value":   types.StringType,
		"comment": types.StringType,
	}
}

func (m *modelRecord) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *modelRecord) FromAPI(_ context.Context, hc hcloud.ZoneRRSetRecord) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Value = types.StringValue(hc.Value)
	if hc.Comment != "" {
		m.Comment = types.StringValue(hc.Comment)
	}

	return diags
}

func (m *modelRecord) ToAPI(_ context.Context) (hc hcloud.ZoneRRSetRecord, diags diag.Diagnostics) {
	hc.Value = m.Value.ValueString()

	if !m.Comment.IsNull() {
		hc.Comment = m.Comment.ValueString()
	}

	return
}

func (m *modelRecord) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func (m *modelRecord) FromTerraform(ctx context.Context, tf types.Object) diag.Diagnostics {
	return tf.As(ctx, m, basetypes.ObjectAsOptions{})
}
