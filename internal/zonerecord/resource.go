package zonerecord

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const ResourceType = "hcloud_zone_record"

var _ resource.Resource = (*Resource)(nil)
var _ resource.ResourceWithConfigure = (*Resource)(nil)
var _ resource.ResourceWithImportState = (*Resource)(nil)
var _ resource.ResourceWithIdentity = (*Resource)(nil)

type Resource struct {
	client *hcloud.Client
}

func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = ResourceType
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = util.MarkdownDescription(`
Provides a Hetzner Cloud Zone Record resource.

This can be used to create, modify, and delete Zone Records.

Managing the TTL, labels and protection level for the Zone Record Set that the Record belongs to is not possible.

Importing this resource is only supported [using an identity](https://developer.hashicorp.com/terraform/plugin/framework/resources/identity#importing-by-identity).

See the [Zone RRSets API documentation](https://docs.hetzner.cloud/reference/cloud#zone-rrsets) for more details.

!> This resource must only be used, when records cannot be managed with a ''hcloud_zone_rrset'' resource.
`)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"zone": schema.StringAttribute{
			MarkdownDescription: "ID or Name of the parent Zone.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Zone Record.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the Zone Record.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"value": schema.StringAttribute{
			MarkdownDescription: "Value of the Zone Record.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"comment": schema.StringAttribute{
			MarkdownDescription: "Comment of the Zone Record.",
			Optional:            true,
			Computed:            true,
		},
	}
}

func (r *Resource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"zone": identityschema.StringAttribute{
				Description:       "ID or Name of the parent Zone.",
				RequiredForImport: true,
			},
			"name": identityschema.StringAttribute{
				Description:       "Name of the Zone Record.",
				RequiredForImport: true,
			},
			"type": identityschema.StringAttribute{
				Description:       "Type of the Zone Record.",
				RequiredForImport: true,
			},
			"value": identityschema.StringAttribute{
				Description:       "Value of the Zone Record.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read request
	var data model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create in API
	rrset := &hcloud.ZoneRRSet{
		Zone: &hcloud.Zone{Name: data.Zone.ValueString()},
		Name: data.Name.ValueString(),
		Type: hcloud.ZoneRRSetType(data.Type.ValueString()),
	}

	opts := hcloud.ZoneRRSetAddRecordsOpts{
		Records: []hcloud.ZoneRRSetRecord{{
			Value: data.Value.ValueString(),
		}},
	}

	if !data.Comment.IsUnknown() && !data.Comment.IsNull() {
		opts.Records[0].Comment = data.Comment.ValueString()
	}

	var actions []*hcloud.Action
	{
		action, _, err := r.client.Zone.AddRRSetRecords(ctx, rrset, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		actions = append(actions, action)

		rrset.Records = opts.Records // Update the state
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, actions...)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write state
	resp.Diagnostics.Append(data.FromAPI(ctx, rrset)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Write identity
	identity := newIdentity(data)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read request
	var identity identityModel
	var data model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	} else {
		identity.FromModel(data)
	}

	// Fetch API
	rrset, recordValue := identity.ToAPI(ctx)

	in, _, err := r.client.Zone.GetRRSetByNameAndType(ctx, rrset.Zone, rrset.Name, rrset.Type)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if in == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	in.Records = slices.DeleteFunc(in.Records, func(o hcloud.ZoneRRSetRecord) bool {
		return o.Value != recordValue
	})
	if len(in.Records) != 1 {
		resp.State.RemoveResource(ctx)
		return
	}

	// Write state
	resp.Diagnostics.Append(data.FromAPI(ctx, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Write identity
	identity.FromModel(data)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read request
	var identity identityModel
	var state, plan model

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	} else {
		identity.FromModel(state)
	}

	// Fetch API
	rrset, recordValue := identity.ToAPI(ctx)

	actions := make([]*hcloud.Action, 0)

	if !plan.Comment.IsUnknown() && !plan.Comment.Equal(state.Comment) {
		opts := hcloud.ZoneRRSetUpdateRecordsOpts{
			Records: []hcloud.ZoneRRSetRecord{
				{Value: recordValue, Comment: plan.Comment.ValueString()},
			},
		}
		action, _, err := r.client.Zone.UpdateRRSetRecords(ctx, rrset, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		actions = append(actions, action)

		rrset.Records = opts.Records // Update the state
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, actions...)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write state
	resp.Diagnostics.Append(state.FromAPI(ctx, rrset)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read state
	var data model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read identity
	var identity identityModel

	if !req.Identity.Raw.IsNull() {
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	} else {
		identity.FromModel(data)
	}

	// Delete in API
	rrset, recordValue := identity.ToAPI(ctx)

	opts := hcloud.ZoneRRSetRemoveRecordsOpts{
		Records: []hcloud.ZoneRRSetRecord{{
			Value: recordValue,
		}},
	}

	action, _, err := r.client.Zone.RemoveRRSetRecords(ctx, rrset, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Fail if import was tried with ID
	if req.ID != "" {
		resp.Diagnostics.AddError(
			"Import with ID not supported.",
			"Using an ID to import hcloud_zone_record resources is not supported. Instead you can use the identity feature to import this resource.",
		)
		return
	}

	// Read request
	var identity identityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	var data model
	resp.Diagnostics.Append(data.FromIdentity(ctx, identity)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
