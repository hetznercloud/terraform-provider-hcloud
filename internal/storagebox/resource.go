package storagebox

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// ResourceType is the type name of the Hetzner Storage Box resource.
const ResourceType = "hcloud_storage_box"

var _ resource.Resource = (*Resource)(nil)
var _ resource.ResourceWithConfigure = (*Resource)(nil)
var _ resource.ResourceWithImportState = (*Resource)(nil)

type Resource struct {
	client *hcloud.Client
}

func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata should return the full name of the resource.
func (r *Resource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = ResourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Resource type. It is separately executed for each
// ReadResource RPC.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema should return the schema for this resource.
func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a Hetzner Storage Box resource.

See the [Storage Box API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-boxes) for more details.
`

	defaultAccessSettings, newDiags := (&modelAccessSettings{
		ReachableExternally: types.BoolValue(false),
		SambaEnabled:        types.BoolValue(false),
		SSHEnabled:          types.BoolValue(false),
		WebDAVEnabled:       types.BoolValue(false),
		ZFSEnabled:          types.BoolValue(false),
	}).ToTerraform(ctx)
	resp.Diagnostics.Append(newDiags...)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box.",
			Computed:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box.",
			Required:            true,
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "Primary username of the Storage Box.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"storage_box_type": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box Type.",
			Required:            true,
		},
		"location": schema.StringAttribute{
			MarkdownDescription: "Name of the Location.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "Password of the Storage Box. Stored in the Terraform state; use `password_wo` to keep it out. Exactly one of `password` and `password_wo` must be set. For more details, see the [Storage Boxes password policy](https://docs.hetzner.cloud/reference/hetzner#storage-boxes-password-policy).",
			Optional:            true,
			Sensitive:           true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(path.MatchRoot("password_wo")),
			},
		},
		"password_wo": schema.StringAttribute{
			MarkdownDescription: "Password of the Storage Box, as a [write-only argument](https://developer.hashicorp.com/terraform/language/resources/ephemeral/write-only): it is never written to the Terraform state. Requires `password_wo_version`. For more details, see the [Storage Boxes password policy](https://docs.hetzner.cloud/reference/hetzner#storage-boxes-password-policy).",
			Optional:            true,
			Sensitive:           true,
			WriteOnly:           true,
			Validators: []validator.String{
				stringvalidator.AlsoRequires(path.MatchRoot("password_wo_version")),
			},
		},
		"password_wo_version": schema.Int64Attribute{
			MarkdownDescription: "Version of `password_wo`. The value of `password_wo` cannot be compared against the API or the state, so a password change is triggered by incrementing this instead.",
			Optional:            true,
			Validators: []validator.Int64{
				int64validator.AlsoRequires(path.MatchRoot("password_wo")),
			},
		},
		"labels": resourceutil.LabelsSchema(),
		"ssh_keys": schema.SetAttribute{
			MarkdownDescription: "SSH public keys in OpenSSH format to inject into the Storage Box. It is not possible to update the SSH Keys through the API, so changing this attribute forces a replace of the Storage Box.",
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.RequiresReplace(),
			},
		},
		"access_settings": schema.SingleNestedAttribute{
			MarkdownDescription: "Access settings of the Storage Box.",
			Optional:            true,
			Computed:            true,
			Default:             objectdefault.StaticValue(defaultAccessSettings),
			Attributes: map[string]schema.Attribute{
				"reachable_externally": schema.BoolAttribute{
					MarkdownDescription: "Whether access from outside the Hetzner network is allowed.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				"samba_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the Samba subsystem is enabled.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				"ssh_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the SSH subsystem is enabled.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				"webdav_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the WebDAV subsystem is enabled.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				"zfs_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the ZFS snapshot folder is visible.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
			},
		},
		"server": schema.StringAttribute{
			MarkdownDescription: "FQDN of the Storage Box.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"system": schema.StringAttribute{
			MarkdownDescription: "Host system of the Storage Box.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"delete_protection": schema.BoolAttribute{
			MarkdownDescription: "Prevent the Storage Box from being accidentally deleted outside of Terraform.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"snapshot_plan": schema.SingleNestedAttribute{
			MarkdownDescription: "Details of the active snapshot plan.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"max_snapshots": schema.Int32Attribute{
					MarkdownDescription: "Maximum amount of Snapshots that will be created by this Snapshot Plan. Older Snapshots will be deleted.",
					Required:            true,
				},
				"minute": schema.Int32Attribute{
					MarkdownDescription: "Minute when the Snapshot Plan is executed (UTC).",
					Required:            true,
				},
				"hour": schema.Int32Attribute{
					MarkdownDescription: "Hour when the Snapshot Plan is executed (UTC).",
					Required:            true,
				},
				"day_of_week": schema.Int32Attribute{
					MarkdownDescription: "Day of the week when the Snapshot Plan is executed. Starts at 0 for Sunday til 6 for Saturday. Note that this differs from the API, which uses 1 (Monday) through 7 (Sunday). Null means every day.",
					Optional:            true,
				},
				"day_of_month": schema.Int32Attribute{
					MarkdownDescription: "Day of the month when the Snapshot Plan is executed. Null means every day.",
					Optional:            true,
				},
			},
		},
	}
}

type resourceModel struct {
	commonModel

	Password types.String `tfsdk:"password"`
	// Write-only attributes are null in plan and state; their value is only
	// readable from the configuration.
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int64  `tfsdk:"password_wo_version"`
	SSHKeys           types.Set    `tfsdk:"ssh_keys"`
}

// password returns the password to send to the API, from whichever of the two
// attributes carries it, and whether one is set at all.
func (m *resourceModel) password(config *resourceModel) (string, bool) {
	if !m.Password.IsNull() && !m.Password.IsUnknown() {
		return m.Password.ValueString(), true
	}
	if config != nil && !config.PasswordWO.IsNull() && !config.PasswordWO.IsUnknown() {
		return config.PasswordWO.ValueString(), true
	}
	return "", false
}

var _ util.ModelFromAPI[*hcloud.StorageBox] = &resourceModel{} // reuse commonModel, as the fields from resourceModel are not readable anyway
var _ util.ModelToTerraform[types.Object] = &resourceModel{}

func (m *resourceModel) tfAttributesTypes() map[string]attr.Type {
	return merge.Maps(
		(&commonModel{}).tfAttributesTypes(),
		map[string]attr.Type{
			"password":            types.StringType,
			"password_wo":         types.StringType,
			"password_wo_version": types.Int64Type,
			"ssh_keys":            types.SetType{ElemType: types.StringType},
		},
	)
}

func (m *resourceModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data, config resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// A write-only attribute is null in the plan and in the state; the configuration
	// is the only place its value can be read from.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password, ok := data.password(&config)
	if !ok {
		resp.Diagnostics.AddError(
			"Missing password",
			"Exactly one of `password` and `password_wo` must be set, and neither carries a value.",
		)
		return
	}

	opts := hcloud.StorageBoxCreateOpts{
		Name:           data.Name.ValueString(),
		StorageBoxType: &hcloud.StorageBoxType{Name: data.StorageBoxType.ValueString()},
		Location:       &hcloud.Location{Name: data.Location.ValueString()},
		Password:       password,
	}

	resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, data.Labels, &opts.Labels)...)

	if !data.SSHKeys.IsUnknown() && !data.SSHKeys.IsNull() {
		sshKeys := make([]string, 0, len(data.SSHKeys.Elements()))

		data.SSHKeys.ElementsAs(ctx, &sshKeys, false)

		for _, sshKey := range sshKeys {
			opts.SSHKeys = append(opts.SSHKeys, &hcloud.SSHKey{PublicKey: sshKey})
		}
	}

	if !data.AccessSettings.IsUnknown() && !data.AccessSettings.IsNull() {
		m := modelAccessSettings{}
		resp.Diagnostics.Append(m.FromTerraform(ctx, data.AccessSettings)...)

		hc, diags := m.ToAPI(ctx)
		resp.Diagnostics.Append(diags...)

		opts.AccessSettings = &hcloud.StorageBoxCreateOptsAccessSettings{
			ReachableExternally: &hc.ReachableExternally,
			SambaEnabled:        &hc.SambaEnabled,
			SSHEnabled:          &hc.SSHEnabled,
			WebDAVEnabled:       &hc.WebDAVEnabled,
			ZFSEnabled:          &hc.ZFSEnabled,
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create in API
	result, _, err := r.client.StorageBox.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Make sure to save the ID immediately so we can recover if the process stops after
	// this call. Terraform marks the resource as "tainted", so it can be deleted and no
	// surprise "duplicate resource" errors happen.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(result.StorageBox.ID))...)

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, result.Action)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.DeleteProtection.IsUnknown() && !data.DeleteProtection.IsNull() {
		action, _, err := r.client.StorageBox.ChangeProtection(ctx, result.StorageBox, hcloud.StorageBoxChangeProtectionOpts{
			Delete: data.DeleteProtection.ValueBoolPointer(),
		})
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.SnapshotPlan.IsUnknown() && !data.SnapshotPlan.IsNull() {
		values := modelSnapshotPlan{}
		resp.Diagnostics.Append(values.FromTerraform(ctx, data.SnapshotPlan)...)

		hc, diags := values.ToAPI(ctx)
		resp.Diagnostics.Append(diags...)

		opts := hcloud.StorageBoxEnableSnapshotPlanOpts{
			MaxSnapshots: hc.MaxSnapshots,
			Minute:       hc.Minute,
			Hour:         hc.Hour,
			DayOfWeek:    hc.DayOfWeek,
			DayOfMonth:   hc.DayOfMonth,
		}

		action, _, err := r.client.StorageBox.EnableSnapshotPlan(ctx, result.StorageBox, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Fetch fresh data from the API
	in, _, err := r.client.StorageBox.GetByID(ctx, result.StorageBox.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, _, err := r.client.StorageBox.GetByID(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if in == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, plan, config resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A stored password is compared; a write-only one cannot be, so its version is
	// what a change is read off.
	resetPassword := !plan.Password.IsUnknown() && !plan.Password.IsNull() && !plan.Password.Equal(data.Password)
	if !config.PasswordWO.IsNull() && !plan.PasswordWOVersion.Equal(data.PasswordWOVersion) {
		resetPassword = true
	}

	storageBox := &hcloud.StorageBox{ID: data.ID.ValueInt64()}

	// Run Actions

	// Action: Delete Protection
	if !plan.DeleteProtection.IsUnknown() && !plan.DeleteProtection.Equal(data.DeleteProtection) {
		action, _, err := r.client.StorageBox.ChangeProtection(ctx, storageBox, hcloud.StorageBoxChangeProtectionOpts{
			Delete: plan.DeleteProtection.ValueBoolPointer(),
		})
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Action: Enable/Disable Snapshot Plan
	if !plan.SnapshotPlan.IsUnknown() && !plan.SnapshotPlan.Equal(data.SnapshotPlan) {
		if plan.SnapshotPlan.IsNull() {
			action, _, err := r.client.StorageBox.DisableSnapshotPlan(ctx, storageBox)
			if err != nil {
				resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
				return
			}

			resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			values := modelSnapshotPlan{}
			resp.Diagnostics.Append(values.FromTerraform(ctx, plan.SnapshotPlan)...)

			hc, diags := values.ToAPI(ctx)
			resp.Diagnostics.Append(diags...)

			// Check for errors before sending data to API
			if resp.Diagnostics.HasError() {
				return
			}

			opts := hcloud.StorageBoxEnableSnapshotPlanOpts{
				MaxSnapshots: hc.MaxSnapshots,
				Minute:       hc.Minute,
				Hour:         hc.Hour,
				DayOfWeek:    hc.DayOfWeek,
				DayOfMonth:   hc.DayOfMonth,
			}

			action, _, err := r.client.StorageBox.EnableSnapshotPlan(ctx, storageBox, opts)
			if err != nil {
				resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
				return
			}

			resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	// Action: Change Type
	if !plan.StorageBoxType.IsUnknown() && !plan.StorageBoxType.Equal(data.StorageBoxType) {
		opts := hcloud.StorageBoxChangeTypeOpts{
			StorageBoxType: &hcloud.StorageBoxType{Name: plan.StorageBoxType.ValueString()},
		}

		action, _, err := r.client.StorageBox.ChangeType(ctx, storageBox, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Action: Reset Password
	if resetPassword {
		password, ok := plan.password(&config)
		if !ok {
			resp.Diagnostics.AddError(
				"Missing password",
				"Exactly one of `password` and `password_wo` must be set, and neither carries a value.",
			)
			return
		}

		opts := hcloud.StorageBoxResetPasswordOpts{
			Password: password,
		}

		action, _, err := r.client.StorageBox.ResetPassword(ctx, storageBox, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Action: Update Access Settings
	if !plan.AccessSettings.IsUnknown() && !plan.AccessSettings.Equal(data.AccessSettings) {
		values := modelAccessSettings{}
		resp.Diagnostics.Append(values.FromTerraform(ctx, plan.AccessSettings)...)

		opts := hcloud.StorageBoxUpdateAccessSettingsOpts{
			ReachableExternally: values.ReachableExternally.ValueBoolPointer(),
			SambaEnabled:        values.SambaEnabled.ValueBoolPointer(),
			SSHEnabled:          values.SSHEnabled.ValueBoolPointer(),
			WebDAVEnabled:       values.WebDAVEnabled.ValueBoolPointer(),
			ZFSEnabled:          values.ZFSEnabled.ValueBoolPointer(),
		}

		action, _, err := r.client.StorageBox.UpdateAccessSettings(ctx, storageBox, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update fields on resource
	opts := hcloud.StorageBoxUpdateOpts{}

	if !plan.Name.IsUnknown() && !plan.Name.Equal(data.Name) {
		opts.Name = plan.Name.ValueString()
	}

	if !plan.Labels.IsUnknown() && !plan.Labels.Equal(data.Labels) {
		resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, plan.Labels, &opts.Labels)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Always perform the update call last, even when empty, to populate the state with fresh data returned by
	// the update.
	in, _, err := r.client.StorageBox.Update(ctx, storageBox, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Write data to state
	resp.Diagnostics.Append(data.FromAPI(ctx, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// At this point the change password action was successful.
	// We have to update the values saved in the state, this does not happen in `data.FromAPI()`.
	// A null `password` is written through as well: that is what drops a stored
	// password from the state when a configuration moves over to `password_wo`.
	if !plan.Password.IsUnknown() {
		data.Password = plan.Password
	}
	data.PasswordWOVersion = plan.PasswordWOVersion

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageBox := &hcloud.StorageBox{ID: data.ID.ValueInt64()}

	// Disable delete protection before deleting.
	if !data.DeleteProtection.IsUnknown() && !data.DeleteProtection.IsNull() && data.DeleteProtection.ValueBool() {
		action, _, err := r.client.StorageBox.ChangeProtection(ctx, storageBox, hcloud.StorageBoxChangeProtectionOpts{
			Delete: new(false),
		})
		if err != nil {
			if hcloudutil.APIErrorIsNotFound(err) {
				// Resource does not exist anymore, all good.
				return
			}

			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	result, _, err := r.client.StorageBox.Delete(ctx, storageBox)
	if err != nil {
		if hcloudutil.APIErrorIsNotFound(err) {
			return
		}

		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, result.Action)...)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Not setting the `ssh_keys` value during import will trigger a replacement of the resource, even if the
	// user configured `lifecycle { ignore_changes = [ssh_keys]}`.

	if id, err := strconv.ParseInt(req.ID, 10, 64); err == nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ssh_keys"), []string{})...)
		return
	}

	in, _, err := r.client.StorageBox.GetByName(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if in == nil {
		resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box", "name", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), in.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ssh_keys"), []string{})...)
}
