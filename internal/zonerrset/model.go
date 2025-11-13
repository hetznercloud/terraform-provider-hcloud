package zonerrset

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
	Zone             types.String `tfsdk:"zone"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	TTL              types.Int32  `tfsdk:"ttl"`
	Labels           types.Map    `tfsdk:"labels"`
	ChangeProtection types.Bool   `tfsdk:"change_protection"`
	Records          types.Set    `tfsdk:"records"`
}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"zone":              types.StringType,
		"id":                types.StringType,
		"name":              types.StringType,
		"type":              types.StringType,
		"ttl":               types.Int32Type,
		"labels":            types.MapType{ElemType: types.StringType},
		"change_protection": types.BoolType,
		"records":           types.SetType{ElemType: (&modelRecord{}).tfType()},
	}
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.ZoneRRSet] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.ZoneRRSet) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.StringValue(hc.ID)
	m.Name = types.StringValue(hc.Name)
	m.Type = types.StringValue(string(hc.Type))

	if hc.TTL != nil {
		m.TTL = types.Int32Value(int32(*hc.TTL)) // nolint: gosec
	} else {
		m.TTL = types.Int32Null()
	}

	m.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, hc.Labels)
	diags.Append(newDiags...)

	m.ChangeProtection = types.BoolValue(hc.Protection.Change)

	{
		value := modelRecords{}
		diags.Append(value.FromAPI(ctx, hc.Records)...)

		m.Records, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

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

type modelRecords []modelRecord

var _ util.ModelFromAPI[[]hcloud.ZoneRRSetRecord] = &modelRecords{}
var _ util.ModelFromTerraform[types.Set] = &modelRecords{}
var _ util.ModelToAPI[[]hcloud.ZoneRRSetRecord] = &modelRecords{}
var _ util.ModelToTerraform[types.Set] = &modelRecords{}

func (m *modelRecords) FromAPI(ctx context.Context, hcItems []hcloud.ZoneRRSetRecord) (diags diag.Diagnostics) {
	*m = make([]modelRecord, 0, len(hcItems))
	for _, hcItem := range hcItems {
		value := modelRecord{}
		diags.Append(value.FromAPI(ctx, hcItem)...)

		*m = append(*m, value)
	}

	return diags
}

func (m *modelRecords) ToAPI(ctx context.Context) (hcItems []hcloud.ZoneRRSetRecord, diags diag.Diagnostics) {
	hcItems = make([]hcloud.ZoneRRSetRecord, 0, len(*m))
	for _, value := range *m {
		hcItem, newDiags := value.ToAPI(ctx)
		diags.Append(newDiags...)

		hcItems = append(hcItems, hcItem)
	}

	return hcItems, diags

}

func (m *modelRecords) FromTerraform(ctx context.Context, tf types.Set) diag.Diagnostics {
	*m = make(modelRecords, 0, len(tf.Elements()))
	return tf.ElementsAs(ctx, m, false)
}

func (m *modelRecords) ToTerraform(ctx context.Context) (tf types.Set, diags diag.Diagnostics) {
	tfItems := make([]attr.Value, 0, len(*m))
	for _, value := range *m {
		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
	}

	tf, newDiags := types.SetValue((&modelRecord{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return tf, diags
}
