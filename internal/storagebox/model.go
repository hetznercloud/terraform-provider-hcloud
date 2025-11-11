package storagebox

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

type commonModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Username         types.String `tfsdk:"username"`
	StorageBoxType   types.String `tfsdk:"storage_box_type"`
	Location         types.String `tfsdk:"location"`
	AccessSettings   types.Object `tfsdk:"access_settings"`
	Server           types.String `tfsdk:"server"`
	System           types.String `tfsdk:"system"`
	Labels           types.Map    `tfsdk:"labels"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	SnapshotPlan     types.Object `tfsdk:"snapshot_plan"`

	// Omitted for Resource: status, stats, created
}

var _ util.ModelFromAPI[*hcloud.StorageBox] = &commonModel{}
var _ util.ModelToTerraform[types.Object] = &commonModel{}

func (m *commonModel) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.Int64Type,
		"name":              types.StringType,
		"username":          types.StringType,
		"storage_box_type":  types.StringType,
		"location":          types.StringType,
		"access_settings":   types.ObjectType{AttrTypes: (&modelAccessSettings{}).tfAttributesTypes()},
		"server":            types.StringType,
		"system":            types.StringType,
		"labels":            types.MapType{ElemType: types.StringType},
		"delete_protection": types.BoolType,
		"snapshot_plan":     types.ObjectType{AttrTypes: (&modelSnapshotPlan{}).tfAttributesTypes()},
	}
}

func (m *commonModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *commonModel) FromAPI(ctx context.Context, hc *hcloud.StorageBox) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	m.ID = types.Int64Value(hc.ID)
	m.Name = types.StringValue(hc.Name)
	m.Username = types.StringValue(hc.Username)
	m.StorageBoxType = types.StringValue(hc.StorageBoxType.Name)
	m.Location = types.StringValue(hc.Location.Name)

	{
		value := modelAccessSettings{}
		diags.Append(value.FromAPI(ctx, hc.AccessSettings)...)

		m.AccessSettings, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	m.Server = types.StringValue(hc.Server)
	m.System = types.StringValue(hc.System)

	m.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, hc.Labels)
	diags.Append(newDiags...)

	m.DeleteProtection = types.BoolValue(hc.Protection.Delete)

	if hc.SnapshotPlan != nil {
		value := modelSnapshotPlan{}
		diags.Append(value.FromAPI(ctx, hc.SnapshotPlan)...)

		m.SnapshotPlan, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	} else {
		m.SnapshotPlan = types.ObjectNull((&modelSnapshotPlan{}).tfAttributesTypes())
	}

	return diags
}

func (m *commonModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type modelAccessSettings struct {
	ReachableExternally types.Bool `tfsdk:"reachable_externally"`
	SambaEnabled        types.Bool `tfsdk:"samba_enabled"`
	SSHEnabled          types.Bool `tfsdk:"ssh_enabled"`
	WebDAVEnabled       types.Bool `tfsdk:"webdav_enabled"`
	ZFSEnabled          types.Bool `tfsdk:"zfs_enabled"`
}

var _ util.ModelFromAPI[hcloud.StorageBoxAccessSettings] = &modelAccessSettings{}
var _ util.ModelFromTerraform[types.Object] = &modelAccessSettings{}
var _ util.ModelToAPI[hcloud.StorageBoxAccessSettings] = &modelAccessSettings{}
var _ util.ModelToTerraform[types.Object] = &modelAccessSettings{}

func (m *modelAccessSettings) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"reachable_externally": types.BoolType,
		"samba_enabled":        types.BoolType,
		"ssh_enabled":          types.BoolType,
		"webdav_enabled":       types.BoolType,
		"zfs_enabled":          types.BoolType,
	}
}

func (m *modelAccessSettings) FromAPI(_ context.Context, hc hcloud.StorageBoxAccessSettings) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ReachableExternally = types.BoolValue(hc.ReachableExternally)
	m.SambaEnabled = types.BoolValue(hc.SambaEnabled)
	m.SSHEnabled = types.BoolValue(hc.SSHEnabled)
	m.WebDAVEnabled = types.BoolValue(hc.WebDAVEnabled)
	m.ZFSEnabled = types.BoolValue(hc.ZFSEnabled)

	return diags
}

func (m *modelAccessSettings) FromTerraform(ctx context.Context, tf types.Object) diag.Diagnostics {
	return tf.As(ctx, m, basetypes.ObjectAsOptions{})
}

func (m *modelAccessSettings) ToAPI(_ context.Context) (hc hcloud.StorageBoxAccessSettings, diags diag.Diagnostics) {
	hc.ReachableExternally = m.ReachableExternally.ValueBool()
	hc.SambaEnabled = m.SambaEnabled.ValueBool()
	hc.SSHEnabled = m.SSHEnabled.ValueBool()
	hc.WebDAVEnabled = m.WebDAVEnabled.ValueBool()
	hc.ZFSEnabled = m.ZFSEnabled.ValueBool()

	return
}

func (m *modelAccessSettings) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

type modelSnapshotPlan struct {
	MaxSnapshots types.Int32 `tfsdk:"max_snapshots"`
	Minute       types.Int32 `tfsdk:"minute"`
	Hour         types.Int32 `tfsdk:"hour"`
	DayOfWeek    types.Int32 `tfsdk:"day_of_week"`
	DayOfMonth   types.Int32 `tfsdk:"day_of_month"`
}

var _ util.ModelFromAPI[*hcloud.StorageBoxSnapshotPlan] = &modelSnapshotPlan{}
var _ util.ModelFromTerraform[types.Object] = &modelSnapshotPlan{}
var _ util.ModelToAPI[*hcloud.StorageBoxSnapshotPlan] = &modelSnapshotPlan{}
var _ util.ModelToTerraform[types.Object] = &modelSnapshotPlan{}

func (m *modelSnapshotPlan) tfAttributesTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"max_snapshots": types.Int32Type,
		"minute":        types.Int32Type,
		"hour":          types.Int32Type,
		"day_of_week":   types.Int32Type,
		"day_of_month":  types.Int32Type,
	}
}

func (m *modelSnapshotPlan) FromAPI(_ context.Context, hc *hcloud.StorageBoxSnapshotPlan) diag.Diagnostics {
	var diags diag.Diagnostics

	m.MaxSnapshots = types.Int32Value(int32(hc.MaxSnapshots)) //nolint:gosec
	m.Minute = types.Int32Value(int32(hc.Minute))             //nolint:gosec
	m.Hour = types.Int32Value(int32(hc.Hour))                 //nolint:gosec

	if hc.DayOfWeek != nil {
		m.DayOfWeek = types.Int32Value(int32(*hc.DayOfWeek)) //nolint:gosec
	} else {
		m.DayOfWeek = types.Int32Null()
	}

	if hc.DayOfMonth != nil {
		m.DayOfMonth = types.Int32Value(int32(*hc.DayOfMonth)) //nolint:gosec
	} else {
		m.DayOfMonth = types.Int32Null()
	}

	return diags
}

func (m *modelSnapshotPlan) FromTerraform(ctx context.Context, tf types.Object) diag.Diagnostics {
	return tf.As(ctx, m, basetypes.ObjectAsOptions{})
}

func (m *modelSnapshotPlan) ToAPI(_ context.Context) (hc *hcloud.StorageBoxSnapshotPlan, diags diag.Diagnostics) {
	hc = &hcloud.StorageBoxSnapshotPlan{}

	hc.MaxSnapshots = int(m.MaxSnapshots.ValueInt32())
	hc.Minute = int(m.Minute.ValueInt32())
	hc.Hour = int(m.Hour.ValueInt32())

	if !m.DayOfWeek.IsNull() {
		hc.DayOfWeek = hcloud.Ptr(time.Weekday(int(m.DayOfWeek.ValueInt32())))
	}

	if !m.DayOfMonth.IsNull() {
		hc.DayOfMonth = hcloud.Ptr(int(m.DayOfMonth.ValueInt32()))
	}

	return
}

func (m *modelSnapshotPlan) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}
