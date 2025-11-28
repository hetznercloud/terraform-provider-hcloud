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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/convutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const ResourceType = "hcloud_zone_record"

var _ resource.Resource = (*ZoneRecordResource)(nil)
var _ resource.ResourceWithConfigure = (*ZoneRecordResource)(nil)
var _ resource.ResourceWithImportState = (*ZoneRecordResource)(nil)
var _ resource.ResourceWithIdentity = (*ZoneRecordResource)(nil)

type ZoneRecordResource struct {
	client *hcloud.Client
}

func NewResource() resource.Resource {
	return &ZoneRecordResource{}
}

func (r *ZoneRecordResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = ResourceType
}

func (r *ZoneRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	experimental.DNS.AppendDiagnostic(&resp.Diagnostics)

	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ZoneRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a Hetzner Cloud Zone Record resource.

This can be used to create, modify, and delete Zone Records.

It is not possible to set the Labels and Protection level for the RRSet that the Record belongs to.

See the [Zone RRSets API documentation](https://docs.hetzner.cloud/reference/cloud#zone-rrsets) for more details.
`

	experimental.DNS.AppendNotice(&resp.Schema.MarkdownDescription)

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
		"ttl": schema.Int32Attribute{
			MarkdownDescription: "Time To Live (TTL) of the Zone Record. All records with the same Name and Type must have the same TTL.",
			Optional:            true,
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
		},
	}
}

func (r *ZoneRecordResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
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

func (r *ZoneRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	}

	if !data.TTL.IsUnknown() {
		opts := hcloud.ZoneRRSetChangeTTLOpts{
			TTL: convutil.Int32ToInt(data.TTL.ValueInt32Pointer()),
		}
		action, _, err := r.client.Zone.ChangeRRSetTTL(ctx, rrset, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		actions = append(actions, action)
	}

	if err := r.client.Action.WaitFor(ctx, actions...); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Write state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// Write identity
	identity := newIdentity(data)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
}

func (r *ZoneRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

func (r *ZoneRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	rrset, _ := identity.ToAPI(ctx)

	actions := make([]*hcloud.Action, 0)

	if !plan.TTL.IsUnknown() && !plan.TTL.Equal(state.TTL) {
		opts := hcloud.ZoneRRSetChangeTTLOpts{
			TTL: convutil.Int32ToInt(plan.TTL.ValueInt32Pointer()),
		}
		action, _, err := r.client.Zone.ChangeRRSetTTL(ctx, rrset, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		actions = append(actions, action)
	}

	if !plan.Comment.IsUnknown() && !plan.Comment.Equal(state.Comment) {
		// TODO: update comment

		// opts := hcloud.ZoneRRSetChangeTTLOpts{
		// 	TTL: convutil.Int32ToInt(plan.TTL.ValueInt32Pointer()),
		// }
		// action, _, err := r.client.Zone.ChangeRRSetTTL(ctx, rrset, opts)
		// if err != nil {
		// 	resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		// 	return
		// }

		// actions = append(actions, action)
	}

	if err := r.client.Action.WaitFor(ctx, actions...); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ZoneRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	if err = r.client.Action.WaitFor(ctx, action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
}

func (r *ZoneRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
