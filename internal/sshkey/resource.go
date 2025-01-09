package sshkey

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// ResourceType is the type name of the Hetzner Cloud SSH Key resource.
const ResourceType = "hcloud_ssh_key"

var _ resource.Resource = (*resourceImpl)(nil)
var _ resource.ResourceWithConfigure = (*resourceImpl)(nil)
var _ resource.ResourceWithImportState = (*resourceImpl)(nil)

type resourceImpl struct {
	client *hcloud.Client
}

func NewResource() resource.Resource {
	return &resourceImpl{}
}

// Metadata should return the full name of the data source.
func (r *resourceImpl) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = ResourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (r *resourceImpl) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceImpl) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a Hetzner Cloud SSH Key resource to manage SSH Keys for server access.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the SSH Key.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the SSH Key.",
			Required:            true,
		},
		"public_key": schema.StringAttribute{
			MarkdownDescription: "Public key of the SSH Key pair. If this is a file, it can be read using the `file` interpolation function.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"fingerprint": schema.StringAttribute{
			MarkdownDescription: "Fingerprint of the SSH public key.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"labels": resourceutil.LabelsSchema(),
	}
}

func (r *resourceImpl) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := hcloud.SSHKeyCreateOpts{
		Name:      data.Name.ValueString(),
		PublicKey: data.PublicKey.ValueString(),
	}
	if !data.Labels.IsNull() {
		hcloudutil.TerraformLabelsToHCloud(ctx, data.Labels, &opts.Labels)
	}

	in, _, err := r.client.SSHKey.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(populateResourceData(ctx, &data, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceImpl) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, newDiags := resourceutil.ParseID(data.ID)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, _, err := r.client.SSHKey.GetByID(ctx, id)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if in == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(populateResourceData(ctx, &data, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceImpl) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, plan resourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := hcloud.SSHKeyUpdateOpts{}

	if !plan.Name.Equal(data.Name) {
		opts.Name = plan.Name.ValueString()
	}

	if !plan.Labels.Equal(data.Labels) {
		hcloudutil.TerraformLabelsToHCloud(ctx, plan.Labels, &opts.Labels)
	}

	id, newDiags := resourceutil.ParseID(data.ID)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, _, err := r.client.SSHKey.Update(ctx, &hcloud.SSHKey{ID: id}, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(populateResourceData(ctx, &data, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceImpl) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, newDiags := resourceutil.ParseID(data.ID)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.SSHKey.Delete(ctx, &hcloud.SSHKey{ID: id})
	if err != nil {
		if hcloudutil.APIErrorIsNotFound(err) { // SSH key does not exist
			return
		}

		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
}
func (r *resourceImpl) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
