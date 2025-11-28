package storageboxsnapshot

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
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	IsAutomatic  types.Bool   `tfsdk:"is_automatic"`
	Labels       types.Map    `tfsdk:"labels"`
	StorageBoxID types.Int64  `tfsdk:"storage_box_id"`

	// Omitted for Resource: stats, created
}

var _ util.ModelFromAPI[*hcloud.StorageBoxSnapshot] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.Int64Type,
		"name":           types.StringType,
		"description":    types.StringType,
		"is_automatic":   types.BoolType,
		"labels":         types.MapType{ElemType: types.StringType},
		"storage_box_id": types.Int64Type,
	}
}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.StorageBoxSnapshot) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Name = types.StringValue(hc.Name)
	m.Description = types.StringValue(hc.Description)
	m.IsAutomatic = types.BoolValue(hc.IsAutomatic)
	m.StorageBoxID = types.Int64Value(hc.StorageBox.ID)

	m.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, hc.Labels)
	diags.Append(newDiags...)

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type modelStats struct {
	Size           types.Int64 `tfsdk:"size"`
	SizeFilesystem types.Int64 `tfsdk:"size_filesystem"`
}

var _ util.ModelFromAPI[hcloud.StorageBoxSnapshotStats] = &modelStats{}
var _ util.ModelFromTerraform[types.Object] = &modelStats{}
var _ util.ModelToAPI[hcloud.StorageBoxSnapshotStats] = &modelStats{}
var _ util.ModelToTerraform[types.Object] = &modelStats{}

func (m *modelStats) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"size":            types.Int64Type,
		"size_filesystem": types.Int64Type,
	}
}

func (m *modelStats) FromAPI(_ context.Context, hc hcloud.StorageBoxSnapshotStats) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Size = types.Int64Value(int64(hc.Size))                     // nolint:gosec
	m.SizeFilesystem = types.Int64Value(int64(hc.SizeFilesystem)) // nolint:gosec

	return diags
}

func (m *modelStats) FromTerraform(ctx context.Context, tf types.Object) diag.Diagnostics {
	return tf.As(ctx, m, basetypes.ObjectAsOptions{})
}

func (m *modelStats) ToAPI(_ context.Context) (hc hcloud.StorageBoxSnapshotStats, diags diag.Diagnostics) {

	return
}

func (m *modelStats) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
