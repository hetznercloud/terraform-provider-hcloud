package storageboxtype

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/deprecation"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"
)

type model struct {
	ID                     types.Int64  `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	SnapshotLimit          types.Int64  `tfsdk:"snapshot_limit"`
	AutomaticSnapshotLimit types.Int64  `tfsdk:"automatic_snapshot_limit"`
	SubaccountsLimit       types.Int64  `tfsdk:"subaccounts_limit"`
	Size                   types.Int64  `tfsdk:"size"`

	deprecation.DeprecationModel
}

func (m *model) tfAttributesTypes() map[string]attr.Type {
	return merge.Maps(map[string]attr.Type{
		"id":                       types.Int64Type,
		"name":                     types.StringType,
		"description":              types.StringType,
		"snapshot_limit":           types.Int64Type,
		"automatic_snapshot_limit": types.Int64Type,
		"subaccounts_limit":        types.Int64Type,
		"size":                     types.Int64Type,
	}, deprecation.AttrTypes())
}

func (m *model) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

var _ util.ModelFromAPI[*hcloud.StorageBoxType] = &model{}
var _ util.ModelToTerraform[types.Object] = &model{}

func (m *model) FromAPI(ctx context.Context, hc *hcloud.StorageBoxType) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Name = types.StringValue(hc.Name)
	m.Description = types.StringValue(hc.Description)
	m.SnapshotLimit = types.Int64PointerValue(intToInt64Ptr(hc.SnapshotLimit))
	m.AutomaticSnapshotLimit = types.Int64PointerValue(intToInt64Ptr(hc.AutomaticSnapshotLimit))
	m.SubaccountsLimit = types.Int64Value(int64(hc.SubaccountsLimit))
	m.Size = types.Int64Value(hc.Size)

	m.DeprecationModel, newDiags = deprecation.NewDeprecationModel(ctx, hc)
	diags.Append(newDiags...)

	return diags
}

func (m *model) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func intToInt64Ptr(i *int) *int64 {
	if i == nil {
		return nil
	}
	v := int64(*i)
	return &v
}

type listModel struct {
	ID              types.String `tfsdk:"id"`
	StorageBoxTypes types.List   `tfsdk:"storage_box_types"`
}

func (m *listModel) FromAPI(ctx context.Context, hc []*hcloud.StorageBoxType) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	tfIDs := make([]string, 0, len(hc))
	tfItems := make([]attr.Value, 0, len(hc))

	for _, item := range hc {
		tfIDs = append(tfIDs, util.FormatID(item.ID))

		var value model
		diags.Append(value.FromAPI(ctx, item)...)

		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
	}

	m.ID = types.StringValue(datasourceutil.ListID(tfIDs))
	m.StorageBoxTypes, newDiags = types.ListValue((&model{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return diags
}
