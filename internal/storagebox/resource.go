package storagebox

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// ResourceType is the type name of the Hetzner Storage Box resource.
const ResourceType = "hcloud_storage_box"

var _ resource.Resource = (*Resource)(nil)
var _ resource.ResourceWithConfigure = (*Resource)(nil)
var _ resource.ResourceWithImportState = (*Resource)(nil)
var _ resource.ResourceWithConfigValidators = (*Resource)(nil)

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
	experimental.StorageBox.AppendDiagnostic(&resp.Diagnostics)

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

	experimental.StorageBox.AppendNotice(&resp.Schema.MarkdownDescription)

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
			MarkdownDescription: "Password of the Storage Box. For more details, see the [Storage Boxes password policy](https://docs.hetzner.cloud/reference/hetzner#storage-boxes-password-policy).",
			Required:            true,
			Sensitive:           true,
			// TODO: Should we generate a password and make it available to the user? Feels painful to always generate a random password, and it can be reset if lost
			// TODO: Evaluate if it makes sense to set `WriteOnly: true,`, as we can not import the password
		},
		"labels": resourceutil.LabelsSchema(),
		"ssh_keys": schema.ListAttribute{
			MarkdownDescription: "SSH public keys in OpenSSH format to inject into the Storage Box. It is not possible to update the SSH Keys through the API after creating the Storage Box, so changing this attribute will delete and re-create the Storage Box, you can also add the SSH Keys to the Storage Box manually.",
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})), // TODO: is there some easier way to get an empty list?
			PlanModifiers: []planmodifier.List{
				// No way to update them through the API, similar to servers
				listplanmodifier.RequiresReplace(),
			},
			// TODO: Evaluate if it makes sense to set `WriteOnly: true,`, as we can not import the password
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
		},
		"system": schema.StringAttribute{
			MarkdownDescription: "Host system of the Storage Box.",
			Computed:            true,
		},
		"delete_protection": schema.BoolAttribute{
			MarkdownDescription: "Prevent the Storage Box from being accidentally deleted outside of Terraform.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"snapshot_plan": schema.SingleNestedAttribute{
			MarkdownDescription: "",
			Optional:            true,
			// TODO: Default or Computed necessary?
			Attributes: map[string]schema.Attribute{
				"max_snapshots": schema.Int32Attribute{
					MarkdownDescription: util.MarkdownDescription(`
						Maximum amount of Snapshots that will be created by this Snapshot Plan.

						Older Snapshots will be deleted.
					`),
					Required: true,
				},
				"minute": schema.Int32Attribute{
					MarkdownDescription: "Minute when the Snapshot Plan is executed (UTC).",
					Required:            true,
				},
				"hour": schema.Int32Attribute{
					MarkdownDescription: "Hour when the Snapshot Plan is executed (UTC).\n",
					Required:            true,
				},
				"day_of_week": schema.Int32Attribute{
					// TODO: Also accept string days similar to CLI?
					// TODO: Exactly one of day_of_week, day_of_month
					MarkdownDescription: util.MarkdownDescription(`
						Day of the week when the Snapshot Plan is executed.

						Starts at 1 for Monday til 7 for Sunday. Null means every day.
					`),
					Optional: true,
				},
				"day_of_month": schema.Int32Attribute{
					MarkdownDescription: util.MarkdownDescription(`
						Day of the month when the Snapshot Plan is executed.

						Null means every day.
					`),
					Optional: true,
				},
			},
		},
	}
}

type resourceModel struct {
	commonModel

	Password types.String `tfsdk:"password"`
	SSHKeys  types.List   `tfsdk:"ssh_keys"`
}

var _ util.ModelFromAPI[*hcloud.StorageBox] = &resourceModel{} // reuse commonModel, as the fields from resourceModel are not readable anyway
var _ util.ModelToTerraform[types.Object] = &resourceModel{}

func (m *resourceModel) tfAttributesTypes() map[string]attr.Type {
	return merge.Maps(
		(&commonModel{}).tfAttributesTypes(),
		map[string]attr.Type{
			"password": types.StringType,
			"ssh_keys": types.ListType{ElemType: types.StringType},
		},
	)
}

func (m *resourceModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("snapshot_plan").AtName("day_of_week"),
			path.MatchRoot("snapshot_plan").AtName("day_of_month"),
		),
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := hcloud.StorageBoxCreateOpts{
		Name:           data.Name.ValueString(),
		StorageBoxType: &hcloud.StorageBoxType{Name: data.StorageBoxType.ValueString()},
		Location:       &hcloud.Location{Name: data.Location.ValueString()},
		Password:       data.Password.ValueString(),
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

	if err = r.client.Action.WaitFor(ctx, result.Action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// TODO: (for Actions PR)
	//   Set Delete Protection
	//   Enable Snapshot Plan

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
	var data, plan resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageBox := &hcloud.StorageBox{ID: data.ID.ValueInt64()}

	// Run Actions

	// TODO: (for Actions PR)
	//   Set Delete Protection
	//   Enable Snapshot Plan
	//   Change Type
	//   Reset Password
	//   Update Access Settings

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, _, err := r.client.StorageBox.Delete(ctx, &hcloud.StorageBox{ID: data.ID.ValueInt64()})
	if err != nil {
		if hcloudutil.APIErrorIsNotFound(err) {
			return
		}

		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if err = r.client.Action.WaitFor(ctx, result.Action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if id, err := strconv.ParseInt(req.ID, 10, 64); err == nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
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
}
