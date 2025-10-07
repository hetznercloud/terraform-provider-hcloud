package zone

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
	ID                 types.Int64  `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Mode               types.String `tfsdk:"mode"`
	TTL                types.Int32  `tfsdk:"ttl"`
	Labels             types.Map    `tfsdk:"labels"`
	DeleteProtection   types.Bool   `tfsdk:"delete_protection"`
	PrimaryNameservers types.List   `tfsdk:"primary_nameservers"`

	AuthoritativeNameservers types.Object `tfsdk:"authoritative_nameservers"`
	Registrar                types.String `tfsdk:"registrar"`
}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                        types.Int64Type,
		"name":                      types.StringType,
		"mode":                      types.StringType,
		"ttl":                       types.Int32Type,
		"labels":                    types.MapType{ElemType: types.StringType},
		"delete_protection":         types.BoolType,
		"primary_nameservers":       types.ListType{ElemType: (&modelPrimaryNameserver{}).tfType()},
		"authoritative_nameservers": types.ObjectType{AttrTypes: (&modelAuthoritativeNameservers{}).tfAttributesTypes()},
		"registrar":                 types.StringType,
	}
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.Zone] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.Zone) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Name = types.StringValue(hc.Name)
	m.Mode = types.StringValue(string(hc.Mode))
	m.TTL = types.Int32Value(int32(hc.TTL)) // nolint: gosec

	m.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, hc.Labels)
	diags.Append(newDiags...)

	m.DeleteProtection = types.BoolValue(hc.Protection.Delete)

	{
		value := modelPrimaryNameservers{}
		diags.Append(value.FromAPI(ctx, hc.PrimaryNameservers)...)

		m.PrimaryNameservers, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	{
		value := modelAuthoritativeNameservers{}
		diags.Append(value.FromAPI(ctx, hc.AuthoritativeNameservers)...)

		m.AuthoritativeNameservers, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	m.Registrar = types.StringValue(string(hc.Registrar))

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type modelPrimaryNameserver struct {
	Address       types.String `tfsdk:"address"`
	Port          types.Int32  `tfsdk:"port"`
	TSIGAlgorithm types.String `tfsdk:"tsig_algorithm"`
	TSIGKey       types.String `tfsdk:"tsig_key"`
}

var _ util.ModelFromAPI[hcloud.ZonePrimaryNameserver] = &modelPrimaryNameserver{}
var _ util.ModelFromTerraform[types.Object] = &modelPrimaryNameserver{}
var _ util.ModelToAPI[hcloud.ZonePrimaryNameserver] = &modelPrimaryNameserver{}
var _ util.ModelToTerraform[types.Object] = &modelPrimaryNameserver{}

func (m *modelPrimaryNameserver) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address":        types.StringType,
		"port":           types.Int32Type,
		"tsig_algorithm": types.StringType,
		"tsig_key":       types.StringType,
	}
}

func (m *modelPrimaryNameserver) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *modelPrimaryNameserver) FromAPI(_ context.Context, hc hcloud.ZonePrimaryNameserver) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Address = types.StringValue(hc.Address)
	if hc.Port != 0 {
		m.Port = types.Int32Value(int32(hc.Port)) // nolint: gosec
	}
	if hc.TSIGAlgorithm != "" {
		m.TSIGAlgorithm = types.StringValue(string(hc.TSIGAlgorithm))
	}
	if hc.TSIGKey != "" {
		m.TSIGKey = types.StringValue(hc.TSIGKey)
	}

	return diags
}

func (m *modelPrimaryNameserver) ToAPI(_ context.Context) (hc hcloud.ZonePrimaryNameserver, diags diag.Diagnostics) {
	hc.Address = m.Address.ValueString()
	if !m.Port.IsNull() {
		hc.Port = int(m.Port.ValueInt32())
	}
	if !m.TSIGAlgorithm.IsNull() {
		hc.TSIGAlgorithm = hcloud.ZoneTSIGAlgorithm(m.TSIGAlgorithm.ValueString())
	}
	if !m.TSIGKey.IsNull() {
		hc.TSIGKey = m.TSIGKey.ValueString()
	}

	return
}

func (m *modelPrimaryNameserver) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func (m *modelPrimaryNameserver) FromTerraform(ctx context.Context, tf types.Object) diag.Diagnostics {
	return tf.As(ctx, m, basetypes.ObjectAsOptions{})
}

type modelPrimaryNameservers []modelPrimaryNameserver

var _ util.ModelFromAPI[[]hcloud.ZonePrimaryNameserver] = &modelPrimaryNameservers{}
var _ util.ModelFromTerraform[types.List] = &modelPrimaryNameservers{}
var _ util.ModelToTerraform[types.List] = &modelPrimaryNameservers{}

func (m *modelPrimaryNameservers) FromAPI(ctx context.Context, hcItems []hcloud.ZonePrimaryNameserver) (diags diag.Diagnostics) {
	values := make([]modelPrimaryNameserver, 0, len(hcItems))
	for _, hcItem := range hcItems {
		value := modelPrimaryNameserver{}
		diags.Append(value.FromAPI(ctx, hcItem)...)

		values = append(values, value)
	}

	*m = values

	return diags
}

func (m *modelPrimaryNameservers) FromTerraform(ctx context.Context, tf types.List) diag.Diagnostics {
	*m = make(modelPrimaryNameservers, 0, len(tf.Elements()))
	return tf.ElementsAs(ctx, m, false)
}

func (m *modelPrimaryNameservers) ToTerraform(ctx context.Context) (tf types.List, diags diag.Diagnostics) {
	tfItems := make([]attr.Value, 0, len(*m))
	for _, value := range *m {
		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
	}

	tf, newDiags := types.ListValue((&modelPrimaryNameserver{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return tf, diags
}

type modelAuthoritativeNameservers struct {
	Assigned types.List `tfsdk:"assigned"`
}

var _ util.ModelFromAPI[hcloud.ZoneAuthoritativeNameservers] = &modelAuthoritativeNameservers{}
var _ util.ModelToTerraform[types.Object] = &modelAuthoritativeNameservers{}

func (m *modelAuthoritativeNameservers) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"assigned": types.ListType{ElemType: types.StringType},
	}
}

func (m *modelAuthoritativeNameservers) FromAPI(ctx context.Context, hc hcloud.ZoneAuthoritativeNameservers) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.Assigned, newDiags = types.ListValueFrom(ctx, types.StringType, hc.Assigned)
	diags.Append(newDiags...)

	return diags
}

func (m *modelAuthoritativeNameservers) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
