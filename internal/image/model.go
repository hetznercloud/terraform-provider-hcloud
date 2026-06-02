package image

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

type model struct {
	ID           types.Int64  `tfsdk:"id"`
	Type         types.String `tfsdk:"type"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Labels       types.Map    `tfsdk:"labels"`
	Created      types.String `tfsdk:"created"`
	OSFlavor     types.String `tfsdk:"os_flavor"`
	OSVersion    types.String `tfsdk:"os_version"`
	Architecture types.String `tfsdk:"architecture"`
	RapidDeploy  types.Bool   `tfsdk:"rapid_deploy"`
	Deprecated   types.String `tfsdk:"deprecated"`
}

var _ util.ModelFromAPI[*hcloud.Image] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":           types.Int64Type,
		"type":         types.StringType,
		"name":         types.StringType,
		"description":  types.StringType,
		"labels":       types.MapType{ElemType: types.StringType},
		"created":      types.StringType,
		"os_flavor":    types.StringType,
		"os_version":   types.StringType,
		"architecture": types.StringType,
		"rapid_deploy": types.BoolType,
		"deprecated":   types.StringType,
	}
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.Image) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Type = types.StringValue(string(hc.Type))
	m.Name = types.StringValue(hc.Name)
	m.Description = types.StringValue(hc.Description)
	m.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, hc.Labels)
	diags.Append(newDiags...)
	m.Created = types.StringValue(hc.Created.Format(time.RFC3339))
	m.OSFlavor = types.StringValue(hc.OSFlavor)
	m.OSVersion = types.StringValue(hc.OSVersion)
	m.Architecture = types.StringValue(string(hc.Architecture))
	m.RapidDeploy = types.BoolValue(hc.RapidDeploy)

	if !hc.Deprecated.IsZero() {
		m.Deprecated = types.StringValue(hc.Deprecated.Format(time.RFC3339))
	} else {
		m.Deprecated = types.StringNull()
	}

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
