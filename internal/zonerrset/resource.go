package zonerrset

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// ResourceType is the type name of the Hetzner Cloud Zone Resource Record Set resource.
const ResourceType = "hcloud_zone_rrset"

var _ resource.Resource = (*Resource)(nil)
var _ resource.ResourceWithConfigure = (*Resource)(nil)
var _ resource.ResourceWithImportState = (*Resource)(nil)

type Resource struct {
	client *hcloud.Client
}

func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata should return the full name of the data source.
func (r *Resource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = ResourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	experimental.DNS.AppendDiagnostic(&resp.Diagnostics)

	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a Hetzner Cloud Zone Resource Record Set (RRSet) resource.

This can be used to create, modify, and delete Zone RRSets.

See the [Zone RRSets API documentation](https://docs.hetzner.cloud/reference/cloud#zone-rrsets) for more details.

**RRSets of type SOA:**

SOA records are created or deleted by the Hetzner Cloud API when creating or deleting
the parent Zone, therefor this Terraform resource will:

- import the RRSet in the state, instead of creating it.
- remove the RRSet from the state, instead of deleting it.
- set the SOA record SERIAL value to 0 before saving it to the state.
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
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Zone RRSet.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Zone RRSet.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the Zone RRSet.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"ttl": schema.Int32Attribute{
			MarkdownDescription: "Time To Live (TTL) of the Zone RRSet.",
			Optional:            true,
		},
		"labels": resourceutil.LabelsSchema(),
		"change_protection": schema.BoolAttribute{
			MarkdownDescription: "Whether change protection is enabled.",
			Optional:            true,
			Computed:            true,
		},
		"records": schema.ListNestedAttribute{
			MarkdownDescription: "Records of the Zone RRSet.",
			Required:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"value": schema.StringAttribute{
						MarkdownDescription: "Value of the record.",
						Required:            true,
					},
					"comment": schema.StringAttribute{
						MarkdownDescription: "Comment of the record.",
						Optional:            true,
					},
				},
			},
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := &hcloud.Zone{Name: data.Zone.ValueString()}

	opts := hcloud.ZoneRRSetCreateOpts{
		Name: data.Name.ValueString(),
		Type: hcloud.ZoneRRSetType(data.Type.ValueString()),
	}
	if !data.TTL.IsUnknown() && !data.TTL.IsNull() {
		opts.TTL = hcloud.Ptr(int(data.TTL.ValueInt32()))
	}
	if !data.Labels.IsUnknown() && !data.Labels.IsNull() {
		hcloudutil.TerraformLabelsToHCloud(ctx, data.Labels, &opts.Labels)
	}
	if !data.Records.IsUnknown() && !data.Records.IsNull() {
		values := modelRecords{}
		resp.Diagnostics.Append(values.FromTerraform(ctx, data.Records)...)

		hcItems, diags := values.ToAPI(ctx)
		resp.Diagnostics.Append(diags...)

		opts.Records = hcItems
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var result hcloud.ZoneRRSetCreateResult
	var err error
	var actions []*hcloud.Action

	// SOA records are managed by the backend
	if data.Type.ValueString() == string(hcloud.ZoneRRSetTypeSOA) {
		resp.Diagnostics.AddWarning(
			"SOA records are managed by the API",
			"Importing the SOA record (managed by the API) in the state.",
		)

		rrset, _, err := r.client.Zone.GetRRSetByNameAndType(ctx, zone, opts.Name, opts.Type)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if rrset == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("zone rrset", "id", fmt.Sprintf("%s/%s", opts.Name, opts.Type)))
		}
		result.RRSet = rrset
	} else {

		result, _, err = r.client.Zone.CreateRRSet(ctx, zone, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
	}

	actions = append(actions, result.Action)

	if !data.ChangeProtection.IsUnknown() && !data.ChangeProtection.IsNull() {
		action, _, err := r.client.Zone.ChangeRRSetProtection(ctx, result.RRSet, hcloud.ZoneRRSetChangeProtectionOpts{
			Change: data.ChangeProtection.ValueBoolPointer(),
		})
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

	// Fetch fresh data from the API
	in, _, err := r.client.Zone.GetRRSetByID(ctx, zone, result.RRSet.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	OverrideRecordsSOASerial(in)

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

	zone := &hcloud.Zone{Name: data.Zone.ValueString()}

	in, _, err := r.client.Zone.GetRRSetByID(ctx, zone, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if in == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	OverrideRecordsSOASerial(in)

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

	rrset := &hcloud.ZoneRRSet{
		Zone: &hcloud.Zone{Name: data.Zone.ValueString()},
		ID:   data.ID.ValueString(),
	}

	// Disable change protection before performing other updates
	if !plan.ChangeProtection.IsUnknown() && !plan.ChangeProtection.Equal(data.ChangeProtection) && !plan.ChangeProtection.ValueBool() {
		action, _, err := r.client.Zone.ChangeRRSetProtection(ctx, rrset, hcloud.ZoneRRSetChangeProtectionOpts{
			Change: plan.ChangeProtection.ValueBoolPointer(),
		})
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		if err := r.client.Action.WaitFor(ctx, action); err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
	}

	actions := make([]*hcloud.Action, 0)

	if !plan.TTL.IsUnknown() && !plan.TTL.Equal(data.TTL) {
		opts := hcloud.ZoneRRSetChangeTTLOpts{}
		if !plan.TTL.IsNull() {
			opts.TTL = hcloud.Ptr(int(plan.TTL.ValueInt32()))
		}

		action, _, err := r.client.Zone.ChangeRRSetTTL(ctx, rrset, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		actions = append(actions, action)
	}

	if !plan.Records.IsUnknown() && !plan.Records.Equal(data.Records) {
		values := modelRecords{}
		resp.Diagnostics.Append(values.FromTerraform(ctx, plan.Records)...)

		hcItems, diags := values.ToAPI(ctx)
		resp.Diagnostics.Append(diags...)

		action, _, err := r.client.Zone.SetRRSetRecords(ctx, rrset, hcloud.ZoneRRSetSetRecordsOpts{
			Records: hcItems,
		})
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

	// Do the update call last, so we can populate the state with fresh data returned by
	// the update.
	opts := hcloud.ZoneRRSetUpdateOpts{}

	if !plan.Labels.IsUnknown() && !plan.Labels.Equal(data.Labels) {
		hcloudutil.TerraformLabelsToHCloud(ctx, plan.Labels, &opts.Labels)
	}

	in, _, err := r.client.Zone.UpdateRRSet(ctx, rrset, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Enable change protection after performing other updates
	if !plan.ChangeProtection.IsUnknown() && !plan.ChangeProtection.Equal(data.ChangeProtection) && plan.ChangeProtection.ValueBool() {
		action, _, err := r.client.Zone.ChangeRRSetProtection(ctx, rrset, hcloud.ZoneRRSetChangeProtectionOpts{
			Change: plan.ChangeProtection.ValueBoolPointer(),
		})
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		if err := r.client.Action.WaitFor(ctx, action); err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		// Update the protection value instead of fetching the API again
		in.Protection.Change = plan.ChangeProtection.ValueBool()
	}

	OverrideRecordsSOASerial(in)

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

	// SOA records are managed by the backend
	if data.Type.ValueString() == string(hcloud.ZoneRRSetTypeSOA) {
		resp.Diagnostics.AddWarning(
			"SOA records are managed by the API",
			"Removing the SOA record (managed by the API) from the state.",
		)
		return
	}

	rrset := &hcloud.ZoneRRSet{
		Zone: &hcloud.Zone{Name: data.Zone.ValueString()},
		ID:   data.ID.ValueString(),
	}

	result, _, err := r.client.Zone.DeleteRRSet(ctx, rrset)
	if err != nil {
		if hcloudutil.APIErrorIsNotFound(err) {
			return
		}

		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if err := r.client.Action.WaitFor(ctx, result.Action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.Append(util.InvalidImportID("$ZONE_ID_OR_NAME/$RRSET_NAME/$RRSET_TYPE", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// OverrideRecordsSOASerial set the serial value to 0 in a SOA Resource Record Set.
func OverrideRecordsSOASerial(hc *hcloud.ZoneRRSet) {
	if hc.Type != hcloud.ZoneRRSetTypeSOA {
		return
	}

	// SOA should have only a single record, this is only defensive
	for i := range hc.Records {
		// hydrogen.ns.hetzner.com. dns.hetzner.com. 2025102142 86400 10800 3600000 3600
		//                                           ^^^^^^^^^^
		parts := strings.Split(hc.Records[i].Value, " ")
		if len(parts) > 2 {
			parts[2] = "0" // Serial
		}
		hc.Records[i].Value = strings.Join(parts, " ")
	}
}
