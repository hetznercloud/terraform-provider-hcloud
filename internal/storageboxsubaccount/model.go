package storageboxsubaccount

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
	ID             types.Int64  `tfsdk:"id"`
	Username       types.String `tfsdk:"username"`
	Description    types.String `tfsdk:"description"`
	HomeDirectory  types.String `tfsdk:"home_directory"`
	Server         types.String `tfsdk:"server"`
	AccessSettings types.Object `tfsdk:"access_settings"`
	Labels         types.Map    `tfsdk:"labels"`
	StorageBoxID   types.Int64  `tfsdk:"storage_box_id"`

	// Omitted for Resource: created
}

var _ util.ModelFromAPI[*hcloud.StorageBoxSubaccount] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.Int64Type,
		"username":        types.StringType,
		"description":     types.StringType,
		"home_directory":  types.StringType,
		"server":          types.StringType,
		"access_settings": types.ObjectType{AttrTypes: (&modelAccessSettings{}).tfAttributesTypes()},
		"labels":          types.MapType{ElemType: types.StringType},
		"storage_box_id":  types.Int64Type,
	}
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.StorageBoxSubaccount) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Username = types.StringValue(hc.Username)
	m.Description = types.StringValue(hc.Description)
	m.HomeDirectory = types.StringValue(hc.HomeDirectory)
	m.Server = types.StringValue(hc.Server)
	m.StorageBoxID = types.Int64Value(hc.StorageBox.ID)

	m.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, hc.Labels)
	diags.Append(newDiags...)

	{
		value := modelAccessSettings{}
		diags.Append(value.FromAPI(ctx, hc.AccessSettings)...)

		m.AccessSettings, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type modelAccessSettings struct {
	ReachableExternally types.Bool `tfsdk:"reachable_externally"`
	SambaEnabled        types.Bool `tfsdk:"samba_enabled"`
	SSHEnabled          types.Bool `tfsdk:"ssh_enabled"`
	WebDAVEnabled       types.Bool `tfsdk:"webdav_enabled"`
	Readonly            types.Bool `tfsdk:"readonly"`
}

var _ util.ModelFromAPI[*hcloud.StorageBoxSubaccountAccessSettings] = &modelAccessSettings{}
var _ util.ModelFromTerraform[types.Object] = &modelAccessSettings{}
var _ util.ModelToTerraform[types.Object] = &modelAccessSettings{}

func (m *modelAccessSettings) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"reachable_externally": types.BoolType,
		"samba_enabled":        types.BoolType,
		"ssh_enabled":          types.BoolType,
		"webdav_enabled":       types.BoolType,
		"readonly":             types.BoolType,
	}
}

func (m *modelAccessSettings) FromAPI(_ context.Context, hc *hcloud.StorageBoxSubaccountAccessSettings) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ReachableExternally = types.BoolValue(hc.ReachableExternally)
	m.SambaEnabled = types.BoolValue(hc.SambaEnabled)
	m.SSHEnabled = types.BoolValue(hc.SSHEnabled)
	m.WebDAVEnabled = types.BoolValue(hc.WebDAVEnabled)
	m.Readonly = types.BoolValue(hc.Readonly)

	return diags
}

func (m *modelAccessSettings) FromTerraform(ctx context.Context, tf types.Object) diag.Diagnostics {
	return tf.As(ctx, m, basetypes.ObjectAsOptions{})
}

func (m *modelAccessSettings) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
