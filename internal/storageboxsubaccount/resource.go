package storageboxsubaccount

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// ResourceType is the type name of the Hetzner Storage Box Subaccount resource.
const ResourceType = "hcloud_storage_box_subaccount"

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
Provides a Hetzner Storage Box Subaccount resource.

See the [Storage Box Subaccounts API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts) for more details.
`

	defaultAccessSettings, newDiags := (&modelAccessSettings{
		ReachableExternally: types.BoolValue(false),
		SambaEnabled:        types.BoolValue(false),
		SSHEnabled:          types.BoolValue(false),
		WebDAVEnabled:       types.BoolValue(false),
		Readonly:            types.BoolValue(false),
	}).ToTerraform(ctx)
	resp.Diagnostics.Append(newDiags...)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"storage_box_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box.",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				int64planmodifier.RequiresReplace(),
			},
		},
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box Subaccount.",
			Computed:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box Subaccount.",
			Optional:            true,
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "A description of the Storage Box Subaccount.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "Username of the Storage Box Subaccount.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"home_directory": schema.StringAttribute{
			MarkdownDescription: "Home directory of the Storage Box Subaccount. The directory will be created if it doesn't exist yet. Must not include a leading slash (`/`).",
			Required:            true,
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "Password of the Storage Box. For more details, see the [Storage Boxes password policy](https://docs.hetzner.cloud/reference/hetzner#storage-boxes-password-policy).",
			Required:            true,
			Sensitive:           true,
		},
		"server": schema.StringAttribute{
			MarkdownDescription: "FQDN of the Storage Box Subaccount.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"access_settings": schema.SingleNestedAttribute{
			MarkdownDescription: "Access settings for the Subaccount.",
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
				"readonly": schema.BoolAttribute{
					MarkdownDescription: "Whether the Subaccount is read-only.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
			},
		},
		"labels": resourceutil.LabelsSchema(),
	}
}

type resourceModel struct {
	model

	Password types.String `tfsdk:"password"`
}

var _ util.ModelFromAPI[*hcloud.StorageBoxSubaccount] = &resourceModel{} // reuse model, as the fields from resourceModel are not readable anyway
var _ util.ModelToTerraform[types.Object] = &resourceModel{}

func (m *resourceModel) tfAttributesTypes() map[string]attr.Type {
	return merge.Maps(
		(&model{}).tfAttributesTypes(),
		map[string]attr.Type{
			"password": types.StringType,
		},
	)
}

func (m *resourceModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageBox := &hcloud.StorageBox{
		ID: data.StorageBoxID.ValueInt64(),
	}

	opts := hcloud.StorageBoxSubaccountCreateOpts{
		Name:          data.Name.ValueString(),
		HomeDirectory: data.HomeDirectory.ValueString(),
		Password:      data.Password.ValueString(),
		Description:   data.Description.ValueString(),
	}

	resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, data.Labels, &opts.Labels)...)

	if !data.AccessSettings.IsUnknown() && !data.AccessSettings.IsNull() {
		m := modelAccessSettings{}
		resp.Diagnostics.Append(m.FromTerraform(ctx, data.AccessSettings)...)

		opts.AccessSettings = &hcloud.StorageBoxSubaccountCreateOptsAccessSettings{
			ReachableExternally: m.ReachableExternally.ValueBoolPointer(),
			SambaEnabled:        m.SambaEnabled.ValueBoolPointer(),
			SSHEnabled:          m.SSHEnabled.ValueBoolPointer(),
			WebDAVEnabled:       m.WebDAVEnabled.ValueBoolPointer(),
			Readonly:            m.Readonly.ValueBoolPointer(),
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create in API
	// For a single storage box, only a single subaccount can be created simultaneously, all others fail with `locked` error.
	var result hcloud.StorageBoxSubaccountCreateResult
	err := control.Retry(2*control.DefaultRetries, func() error {
		var err error

		result, _, err = r.client.StorageBox.CreateSubaccount(ctx, storageBox, opts)
		if err != nil {
			if hcloud.IsError(err,
				hcloud.ErrorCodeLocked,
			) {
				return err
			}

			return control.AbortRetry(err)
		}
		return nil
	})
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Make sure to save the ID immediately so we can recover if the process stops after
	// this call. Terraform marks the resource as "tainted", so it can be deleted and no
	// surprise "duplicate resource" errors happen.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("storage_box_id"), types.Int64Value(storageBox.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(result.Subaccount.ID))...)

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, result.Action)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch fresh data from the API
	in, _, err := r.client.StorageBox.GetSubaccountByID(ctx, storageBox, result.Subaccount.ID)
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

	storageBox := &hcloud.StorageBox{ID: data.StorageBoxID.ValueInt64()}

	in, _, err := r.client.StorageBox.GetSubaccountByID(ctx, storageBox, data.ID.ValueInt64())
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
	var data, plan resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := &hcloud.StorageBoxSubaccount{
		StorageBox: &hcloud.StorageBox{ID: data.StorageBoxID.ValueInt64()},
		ID:         data.ID.ValueInt64(),
	}

	// Run Actions

	// Action: Change Home Directory
	if !plan.HomeDirectory.IsUnknown() && !plan.HomeDirectory.Equal(data.HomeDirectory) {
		action, _, err := r.client.StorageBox.ChangeSubaccountHomeDirectory(ctx, subaccount, hcloud.StorageBoxSubaccountChangeHomeDirectoryOpts{
			HomeDirectory: plan.HomeDirectory.ValueString(),
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

	// Action: Reset Password
	if !plan.Password.IsUnknown() && !plan.Password.Equal(data.Password) {
		action, _, err := r.client.StorageBox.ResetSubaccountPassword(ctx, subaccount, hcloud.StorageBoxSubaccountResetPasswordOpts{
			Password: plan.Password.ValueString(),
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

	// Action: Update Access Settings
	if !plan.AccessSettings.IsUnknown() && !plan.AccessSettings.Equal(data.AccessSettings) {
		m := modelAccessSettings{}
		resp.Diagnostics.Append(m.FromTerraform(ctx, plan.AccessSettings)...)

		opts := hcloud.StorageBoxSubaccountUpdateAccessSettingsOpts{
			ReachableExternally: m.ReachableExternally.ValueBoolPointer(),
			SambaEnabled:        m.SambaEnabled.ValueBoolPointer(),
			SSHEnabled:          m.SSHEnabled.ValueBoolPointer(),
			WebDAVEnabled:       m.WebDAVEnabled.ValueBoolPointer(),
			Readonly:            m.Readonly.ValueBoolPointer(),
		}

		action, _, err := r.client.StorageBox.UpdateSubaccountAccessSettings(ctx, subaccount, opts)
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
	opts := hcloud.StorageBoxSubaccountUpdateOpts{}

	if !data.Name.Equal(plan.Name) {
		opts.Name = plan.Name.ValueString()
	}

	if !data.Description.Equal(plan.Description) {
		opts.Description = plan.Description.ValueStringPointer()
	}

	if !data.Labels.Equal(plan.Labels) {
		resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, plan.Labels, &opts.Labels)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Always perform the update call last, even when empty, to populate the state with fresh data returned by
	// the update.
	in, _, err := r.client.StorageBox.UpdateSubaccount(ctx, subaccount, opts)
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
	// We have to update the value saved in the state, this does not happen in `data.FromAPI()`.
	if !plan.Password.IsUnknown() && !plan.Password.Equal(data.Password) {
		data.Password = plan.Password
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subaccount := &hcloud.StorageBoxSubaccount{
		StorageBox: &hcloud.StorageBox{ID: data.StorageBoxID.ValueInt64()},
		ID:         data.ID.ValueInt64(),
	}

	result, _, err := r.client.StorageBox.DeleteSubaccount(ctx, subaccount)
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
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.Append(util.InvalidImportID("$STORAGE_BOX_ID/$SUBACCOUNT_ID", req.ID))
		return
	}

	storageBoxID, err := util.ParseID(parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse $STORAGE_BOX_ID", fmt.Sprintf("Failed to parse first segment of the import id: %v", err))
	}

	subaccountID, err := util.ParseID(parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse $SUBACCOUNT_ID ", fmt.Sprintf("Failed to parse second segment of the import id: %v", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("storage_box_id"), storageBoxID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), subaccountID)...)
}
