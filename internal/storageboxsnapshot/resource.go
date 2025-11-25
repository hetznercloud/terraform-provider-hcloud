package storageboxsnapshot

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// ResourceType is the type name of the Hetzner Storage Box Snapshot resource.
const ResourceType = "hcloud_storage_box_snapshot"

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
	experimental.StorageBox.AppendDiagnostic(&resp.Diagnostics)

	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema should return the schema for this resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a Hetzner Storage Box Snapshot resource.

See the [Storage Box Snapshots API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots) for more details.
`

	experimental.StorageBox.AppendNotice(&resp.Schema.MarkdownDescription)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"storage_box": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box.",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
				int64planmodifier.RequiresReplace(),
			},
		},
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box Snapshot.",
			Computed:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box Snapshot.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Storage Box Snapshot.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
		},
		"is_automatic": schema.BoolAttribute{
			MarkdownDescription: "Whether the Storage Box Snapshot was created automatically.",
			Computed:            true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"labels": resourceutil.LabelsSchema(),
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageBox := &hcloud.StorageBox{
		ID: data.StorageBoxID.ValueInt64(),
	}

	opts := hcloud.StorageBoxSnapshotCreateOpts{
		Description: data.Description.ValueString(),
	}

	resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, data.Labels, &opts.Labels)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create in API
	result, _, err := r.client.StorageBox.CreateSnapshot(ctx, storageBox, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Make sure to save the ID immediately so we can recover if the process stops after
	// this call. Terraform marks the resource as "tainted", so it can be deleted and no
	// surprise "duplicate resource" errors happen.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(result.Snapshot.ID))...)

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, result.Action)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch fresh data from the API
	in, _, err := r.client.StorageBox.GetSnapshotByID(ctx, storageBox, result.Snapshot.ID)
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
	var data model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageBox := &hcloud.StorageBox{ID: data.StorageBoxID.ValueInt64()}

	in, _, err := r.client.StorageBox.GetSnapshotByID(ctx, storageBox, data.ID.ValueInt64())
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
	var data, plan model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshot := &hcloud.StorageBoxSnapshot{
		StorageBox: &hcloud.StorageBox{ID: data.StorageBoxID.ValueInt64()},
		ID:         data.ID.ValueInt64(),
	}

	opts := hcloud.StorageBoxSnapshotUpdateOpts{}

	if !data.Description.Equal(plan.Description) {
		opts.Description = plan.Description.ValueStringPointer()
	}

	if !data.Labels.Equal(plan.Labels) {
		resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, plan.Labels, &opts.Labels)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	in, _, err := r.client.StorageBox.UpdateSnapshot(ctx, snapshot, opts)
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
	var data model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshot := &hcloud.StorageBoxSnapshot{
		StorageBox: &hcloud.StorageBox{ID: data.StorageBoxID.ValueInt64()},
		ID:         data.ID.ValueInt64(),
	}

	result, _, err := r.client.StorageBox.DeleteSnapshot(ctx, snapshot)
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
		resp.Diagnostics.Append(util.InvalidImportID("$STORAGE_BOX_ID/$SNAPSHOT_ID", req.ID))
		return
	}

	storageBoxID, err := util.ParseID(parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse $STORAGE_BOX_ID", fmt.Sprintf("Failed to parse first segment of the import id: %v", err))
	}

	snapshotID, err := util.ParseID(parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse $SNAPSHOT_ID ", fmt.Sprintf("Failed to parse second segment of the import id: %v", err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("storage_box"), storageBoxID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), snapshotID)...)
}
