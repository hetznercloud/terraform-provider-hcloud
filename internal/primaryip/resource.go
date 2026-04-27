package primaryip

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

const ResourceType = "hcloud_primary_ip"

var _ resource.Resource = (*Resource)(nil)
var _ resource.ResourceWithConfigure = (*Resource)(nil)
var _ resource.ResourceWithConfigValidators = (*Resource)(nil)
var _ resource.ResourceWithImportState = (*Resource)(nil)

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
Provides a Hetzner Cloud Primary IP resource.

See the [Primary IP API documentation](https://docs.hetzner.cloud/reference/cloud#tag/primary-ips) for more details.

## Deprecations

### ''datacenter'' attribute

The ''datacenter'' attribute is deprecated, use the ''location'' attribute instead.

See our the [API changelog](https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters) for more details.

-> Please upgrade to ''v1.58.0+'' of the provider to avoid issues once the Hetzner Cloud API no longer accepts
and returns the ''datacenter'' attribute. This version of the provider remains backward compatible by preserving
the ''datacenter'' value in the state and by extracting the ''location'' name from the ''datacenter'' attribute when
communicating with the API.
`)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Primary IP.",
			Computed:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Primary IP.",
			Required:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the Primary IP (`ipv4` or `ipv6`).",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"location": schema.StringAttribute{
			MarkdownDescription: "Name of the Location for the Primary IP. See the [Hetzner Docs](https://docs.hetzner.com/cloud/general/locations/#what-locations-are-there) for more details about locations.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"datacenter": schema.StringAttribute{
			MarkdownDescription: "Name of the Datacenter for the Primary IP. See the [Hetzner Docs](https://docs.hetzner.com/cloud/general/locations/#what-datacenters-are-there) for more details about datacenters.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			DeprecationMessage: "The datacenter attribute is deprecated and will be removed after 1 July 2026. Please use the location attribute instead. See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters.",
		},
		"assignee_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the resource the Primary IP should be assigned to.",
			Optional:            true,
			Computed:            true,
		},
		"assignee_type": schema.StringAttribute{
			MarkdownDescription: "Type of the resource the Primary IP should be assigned to.",
			Required:            true,
		},
		"auto_delete": schema.BoolAttribute{
			MarkdownDescription: "Whether auto delete is enabled. Setting `auto_delete` to `false` is recommended, because if a server assigned to the managed ip is getting deleted, it will also delete the primary IP which will break the terraform state.",
			Required:            true,
		},
		"labels": resourceutil.LabelsSchema(),
		"delete_protection": schema.BoolAttribute{
			MarkdownDescription: " Whether delete protection is enabled.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"ip_address": schema.StringAttribute{
			MarkdownDescription: "IP address of the Primary IP.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"ip_network": schema.StringAttribute{
			MarkdownDescription: "IP network of the Primary IP for IPv6 addresses. Only set if `type` is `ipv6`.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("location"),
			path.MatchRoot("datacenter"),
			path.MatchRoot("assignee_id"),
		),
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := hcloud.PrimaryIPCreateOpts{
		Name: data.Name.ValueString(),
		Type: hcloud.PrimaryIPType(data.Type.ValueString()),

		AssigneeType: data.AssigneeType.ValueString(),

		AutoDelete: data.AutoDelete.ValueBoolPointer(),
	}

	resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, data.Labels, &opts.Labels)...)

	switch {
	case !data.Location.IsUnknown() && !data.Location.IsNull():
		opts.Location = data.Location.ValueString()
	case !data.Datacenter.IsUnknown() && !data.Datacenter.IsNull():
		// Backward compatible datacenter argument: datacenter hel1-dc2 => location hel1
		parts := strings.Split(data.Datacenter.ValueString(), "-")
		if len(parts) != 2 {
			resp.Diagnostics.AddAttributeError(
				path.Root("datacenter"),
				"Invalid datacenter name",
				fmt.Sprintf("Datacenter name is not valid, expected format $LOCATION-$DATACENTER, but got: %s", data.Datacenter.ValueString()),
			)
		}
		opts.Location = parts[0]
	case !data.AssigneeID.IsUnknown() && !data.AssigneeID.IsNull():
		opts.AssigneeID = data.AssigneeID.ValueInt64Pointer()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create in API
	result, _, err := r.client.PrimaryIP.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Make sure to save the ID immediately so we can recover if the process stops after
	// this call. Terraform marks the resource as "tainted", so it can be deleted and no
	// surprise "duplicate resource" errors happen.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(result.PrimaryIP.ID))...)

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, result.Action)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.DeleteProtection.IsUnknown() && !data.DeleteProtection.IsNull() && data.DeleteProtection.ValueBool() {
		action, _, err := r.client.PrimaryIP.ChangeProtection(ctx, hcloud.PrimaryIPChangeProtectionOpts{
			ID:     result.PrimaryIP.ID,
			Delete: data.DeleteProtection.ValueBool(),
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

	// Fetch fresh data from the API
	in, _, err := r.client.PrimaryIP.GetByID(ctx, result.PrimaryIP.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// backwards-compatibility: Datacenter deprecation
	//nolint:staticcheck
	if in.Datacenter == nil && data.Datacenter.ValueString() != "" {
		//nolint:staticcheck
		in.Datacenter = &hcloud.Datacenter{Name: data.Datacenter.ValueString()}
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

	in, _, err := r.client.PrimaryIP.GetByID(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if in == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// backwards-compatibility: Datacenter deprecation
	//nolint:staticcheck
	if in.Datacenter == nil && data.Datacenter.ValueString() != "" {
		//nolint:staticcheck
		in.Datacenter = &hcloud.Datacenter{Name: data.Datacenter.ValueString()}
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

	primaryIP := &hcloud.PrimaryIP{ID: data.ID.ValueInt64()}

	// Action: Delete Protection
	if !plan.DeleteProtection.IsUnknown() && !plan.DeleteProtection.Equal(data.DeleteProtection) {
		action, _, err := r.client.PrimaryIP.ChangeProtection(ctx, hcloud.PrimaryIPChangeProtectionOpts{
			ID:     primaryIP.ID,
			Delete: plan.DeleteProtection.ValueBool(),
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

	// Action: Assignee ID
	if !plan.AssigneeID.IsUnknown() && !plan.AssigneeID.Equal(data.AssigneeID) {
		if data.AssigneeID.ValueInt64() == 0 { // This handles assignee_id=null
			action, _, err := r.client.PrimaryIP.Unassign(ctx, primaryIP.ID)
			if err != nil {
				resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
				return
			}

			resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			{
				action, _, err := r.client.PrimaryIP.Unassign(ctx, primaryIP.ID)
				if err != nil {
					resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
					return
				}

				resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
			{
				action, _, err := r.client.PrimaryIP.Assign(ctx, hcloud.PrimaryIPAssignOpts{
					ID:           primaryIP.ID,
					AssigneeID:   plan.AssigneeID.ValueInt64(),
					AssigneeType: "server",
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
		}
	}

	// Update fields on resource
	opts := hcloud.PrimaryIPUpdateOpts{}

	if !plan.Name.IsUnknown() && !plan.Name.Equal(data.Name) {
		opts.Name = plan.Name.ValueString()
	}

	if !plan.Labels.IsUnknown() && !plan.Labels.Equal(data.Labels) {
		// Primary IPs labels are weird, opts.Labels is a pointer to a map.
		var labels map[string]string
		resp.Diagnostics.Append(hcloudutil.TerraformLabelsToHCloud(ctx, plan.Labels, &labels)...)
		opts.Labels = &labels
	}

	if !plan.AutoDelete.IsUnknown() && !plan.AutoDelete.Equal(data.AutoDelete) {
		opts.AutoDelete = plan.AutoDelete.ValueBoolPointer()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Always perform the update call last, even when empty, to populate the state with fresh data returned by
	// the update.
	in, _, err := r.client.PrimaryIP.Update(ctx, primaryIP, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// backwards-compatibility: Datacenter deprecation
	//nolint:staticcheck
	if in.Datacenter == nil && plan.Datacenter.ValueString() != "" {
		//nolint:staticcheck
		in.Datacenter = &hcloud.Datacenter{Name: plan.Datacenter.ValueString()}
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

	primaryIP := &hcloud.PrimaryIP{ID: data.ID.ValueInt64()}

	// Unassign Primary IP before deletion
	if data.AssigneeID.ValueInt64() != 0 {
		server, _, err := r.client.Server.GetByID(ctx, data.AssigneeID.ValueInt64())
		if err == nil && server != nil {
			// The server does not have this primary ip assigned anymore, no need to try to detach it before deleting
			// Workaround for https://github.com/hashicorp/terraform/issues/35568
			if server.PublicNet.IPv4.ID == primaryIP.ID ||
				server.PublicNet.IPv6.ID == primaryIP.ID {

				{ // Power off
					action, _, _ := r.client.Server.Poweroff(ctx, server)
					// No error handling

					resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
					if resp.Diagnostics.HasError() {
						return
					}
				}
				{ // Unassign
					action, _, _ := r.client.PrimaryIP.Unassign(ctx, primaryIP.ID)
					// No error handling, because its possible that the primary IP got
					// already unassigned on server destroy

					resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
					if resp.Diagnostics.HasError() {
						return
					}
				}
				{ // Power on
					action, _, _ := r.client.Server.Poweron(ctx, server)
					// No error handling

					resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
					if resp.Diagnostics.HasError() {
						return
					}
				}
			}
		}
	}

	err := control.Retry(2*control.DefaultRetries, func() error {
		_, err := r.client.PrimaryIP.Delete(ctx, primaryIP)
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// Primary IP was already deleted
			return nil
		}
		if hcloud.IsError(err, hcloud.ErrorCodeProtected) {
			// Primary IP is delete protected
			return control.AbortRetry(err)
		}
		return err
	})
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(util.InvalidImportID("$PRIMARY_IP_ID", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
