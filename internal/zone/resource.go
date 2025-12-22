package zone

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// ResourceType is the type name of the Hetzner Cloud Zone resource.
const ResourceType = "hcloud_zone"

var _ resource.Resource = (*Resource)(nil)
var _ resource.ResourceWithConfigure = (*Resource)(nil)
var _ resource.ResourceWithImportState = (*Resource)(nil)
var _ resource.ResourceWithValidateConfig = (*Resource)(nil)

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
	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a Hetzner Cloud Zone resource.

This can be used to create, modify, and delete Zones.

For Internationalized domain names (IDN), see the ` + "`provider::hcloud::idna`" + ` function.

See the [Zones API documentation](https://docs.hetzner.cloud/reference/cloud#zones) for more details.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Zone.",
			Computed:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Zone.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"mode": schema.StringAttribute{
			MarkdownDescription: "Mode of the Zone.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"ttl": schema.Int32Attribute{
			MarkdownDescription: "Default Time To Live (TTL) of the Zone.",
			Optional:            true,
			Computed:            true,
			Default:             int32default.StaticInt32(3600),
		},
		"labels": resourceutil.LabelsSchema(),
		"delete_protection": schema.BoolAttribute{
			MarkdownDescription: "Whether delete protection is enabled.",
			Optional:            true,
			Computed:            true,
		},
		"primary_nameservers": schema.ListNestedAttribute{
			MarkdownDescription: "Primary nameservers of the Zone. Forbidden when mode is primary and required when mode is secondary.",
			Optional:            true,
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Public IPv4 or IPv6 address of the primary nameserver.",
						Required:            true,
					},
					"port": schema.Int32Attribute{
						MarkdownDescription: "Port of the primary nameserver.",
						Optional:            true,
						Computed:            true,
						Default:             int32default.StaticInt32(53),
					},
					"tsig_algorithm": schema.StringAttribute{
						MarkdownDescription: "Transaction signature (TSIG) algorithm used to generate the TSIG key.",
						Optional:            true,
					},
					"tsig_key": schema.StringAttribute{
						MarkdownDescription: "Transaction signature (TSIG) key",
						Optional:            true,
					},
				},
			},
			Validators: []validator.List{
				listvalidator.SizeAtLeast(1),
			},
		},
		"authoritative_nameservers": schema.SingleNestedAttribute{
			MarkdownDescription: "Authoritative nameservers of the Zone.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"assigned": schema.ListAttribute{
					MarkdownDescription: "Authoritative Hetzner nameservers assigned to the Zone.",
					ElementType:         types.StringType,
					Computed:            true,
				},
			},
		},
		"registrar": schema.StringAttribute{
			MarkdownDescription: "Registrar of the Zone.",
			Computed:            true,
		},
	}
}

func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data model

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch data.Mode.ValueString() {
	case string(hcloud.ZoneModePrimary):
		if !data.PrimaryNameservers.IsUnknown() && !data.PrimaryNameservers.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("primary_nameservers"),
				"Forbidden attribute",
				"This attribute is forbidden when mode is primary.",
			)
		}
	case string(hcloud.ZoneModeSecondary):
		if data.PrimaryNameservers.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("primary_nameservers"),
				"Required attribute",
				"This attribute is required when mode is secondary.",
			)
		}
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := hcloud.ZoneCreateOpts{
		Name: data.Name.ValueString(),
		Mode: hcloud.ZoneMode(data.Mode.ValueString()),
	}
	if !data.TTL.IsUnknown() && !data.TTL.IsNull() {
		opts.TTL = hcloud.Ptr(int(data.TTL.ValueInt32()))
	}

	resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, data.Labels, &opts.Labels)...)

	if !data.PrimaryNameservers.IsUnknown() && !data.PrimaryNameservers.IsNull() {
		m := modelPrimaryNameservers{}
		resp.Diagnostics.Append(m.FromTerraform(ctx, data.PrimaryNameservers)...)

		for _, item := range m {
			hcItem, diags := item.ToAPI(ctx)
			resp.Diagnostics.Append(diags...)

			opts.PrimaryNameservers = append(opts.PrimaryNameservers, hcloud.ZoneCreateOptsPrimaryNameserver{
				Address:       hcItem.Address,
				Port:          hcItem.Port,
				TSIGAlgorithm: hcItem.TSIGAlgorithm,
				TSIGKey:       hcItem.TSIGKey,
			})
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	actions := make([]*hcloud.Action, 0)

	result, _, err := r.client.Zone.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
	// Make sure to save the ID immediately so we can recover if the process stops after
	// this call. Terraform marks the resource as "tainted", so it can be deleted and no
	// surprise "duplicate resource" errors happen.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(result.Zone.ID))...)

	actions = append(actions, result.Action)

	if !data.DeleteProtection.IsUnknown() && !data.DeleteProtection.IsNull() {
		action, _, err := r.client.Zone.ChangeProtection(ctx, result.Zone, hcloud.ZoneChangeProtectionOpts{
			Delete: data.DeleteProtection.ValueBoolPointer(),
		})
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		actions = append(actions, action)
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, actions...)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch fresh data from the API
	in, _, err := r.client.Zone.GetByID(ctx, result.Zone.ID)
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

	in, _, err := r.client.Zone.GetByID(ctx, data.ID.ValueInt64())
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

	zone := &hcloud.Zone{ID: data.ID.ValueInt64()}
	actions := make([]*hcloud.Action, 0)

	if !plan.TTL.IsUnknown() && !plan.TTL.Equal(data.TTL) {
		action, _, err := r.client.Zone.ChangeTTL(ctx, zone, hcloud.ZoneChangeTTLOpts{
			TTL: int(plan.TTL.ValueInt32()),
		})
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		actions = append(actions, action)
	}

	if !plan.PrimaryNameservers.IsUnknown() && !plan.PrimaryNameservers.Equal(data.PrimaryNameservers) {
		values := modelPrimaryNameservers{}
		resp.Diagnostics.Append(values.FromTerraform(ctx, plan.PrimaryNameservers)...)

		opts := hcloud.ZoneChangePrimaryNameserversOpts{
			PrimaryNameservers: make([]hcloud.ZoneChangePrimaryNameserversOptsPrimaryNameserver, 0, len(values)),
		}

		for _, value := range values {
			hcItem, diags := value.ToAPI(ctx)
			resp.Diagnostics.Append(diags...)

			opts.PrimaryNameservers = append(opts.PrimaryNameservers, hcloud.ZoneChangePrimaryNameserversOptsPrimaryNameserver{
				Address:       hcItem.Address,
				Port:          hcItem.Port,
				TSIGAlgorithm: hcItem.TSIGAlgorithm,
				TSIGKey:       hcItem.TSIGKey,
			})
		}

		// If data conversion failed we should abort before sending API requests.
		if resp.Diagnostics.HasError() {
			return
		}

		action, _, err := r.client.Zone.ChangePrimaryNameservers(ctx, zone, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		actions = append(actions, action)
	}

	if !plan.DeleteProtection.IsUnknown() && !plan.DeleteProtection.Equal(data.DeleteProtection) {
		action, _, err := r.client.Zone.ChangeProtection(ctx, zone, hcloud.ZoneChangeProtectionOpts{
			Delete: plan.DeleteProtection.ValueBoolPointer(),
		})
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		actions = append(actions, action)
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, actions...)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := hcloud.ZoneUpdateOpts{}

	if !plan.Labels.IsUnknown() && !plan.Labels.Equal(data.Labels) {
		resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, plan.Labels, &opts.Labels)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Always perform the update call last, even when empty, to populate the state with fresh data returned by
	// the update.
	in, _, err := r.client.Zone.Update(ctx, zone, opts)
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

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data model

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, _, err := r.client.Zone.Delete(ctx, &hcloud.Zone{ID: data.ID.ValueInt64()})
	if err != nil {
		if hcloudutil.APIErrorIsNotFound(err) {
			return
		}

		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, result.Action)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if id, err := strconv.ParseInt(req.ID, 10, 64); err == nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
		return
	}

	in, _, err := r.client.Zone.GetByName(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if in == nil {
		resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("zone", "name", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), in.ID)...)
}
