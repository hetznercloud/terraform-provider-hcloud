package server

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/sliceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/validateutil"
)

const NetworkResourceType = "hcloud_server_network"

var _ resource.Resource = (*NetworkResource)(nil)
var _ resource.ResourceWithConfigure = (*NetworkResource)(nil)
var _ resource.ResourceWithConfigValidators = (*NetworkResource)(nil)
var _ resource.ResourceWithImportState = (*NetworkResource)(nil)
var _ resource.ResourceWithModifyPlan = (*NetworkResource)(nil)

type NetworkResource struct {
	client *hcloud.Client
}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

func (r *NetworkResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = NetworkResourceType
}

func (r *NetworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *NetworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = util.MarkdownDescription(`
Manage the attachment of a Server in a Network in the Hetzner Cloud.
`)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"server_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Server.",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"network_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Network to attach the Server to. Using `subnet_id` is preferred. Required if `subnet_id` is not set. If `subnet_id` or `ip` are not set, the Server will be attached to the last subnet (ordered by `ip_range`).",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"subnet_id": schema.StringAttribute{
			MarkdownDescription: "ID of the Subnet to attach the Server to. Required if `network_id` is not set.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"ip": schema.StringAttribute{
			CustomType:          iptypes.IPAddressType{},
			MarkdownDescription: "IP to assign to the Server.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.String{
				validateutil.IP(),
			},
		},
		"alias_ips": schema.SetAttribute{
			MarkdownDescription: "Additional IPs to assign to the Server.",
			ElementType:         iptypes.IPAddressType{},
			Optional:            true,
			Computed:            true,
			Default:             setdefault.StaticValue(types.SetValueMust(iptypes.IPAddressType{}, nil)),
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(validateutil.IP()),
			},
		},
		"mac_address": schema.StringAttribute{
			MarkdownDescription: "MAC address of the Server on the Network.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

func (r *NetworkResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("network_id"),
			path.MatchRoot("subnet_id"),
		),
	}
}

func (r *NetworkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Do not modify on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not modify on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	var data networkResourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.SubnetID.IsUnknown() && !data.SubnetID.IsNull() {
		subnetNetwork, subnetIPRange, err := r.ParseSubnetID(data.SubnetID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("subnet_id"),
				"Invalid Subnet ID",
				util.TitleCase(err.Error()),
			)
		}

		if !data.NetworkID.IsUnknown() && !data.NetworkID.IsNull() {
			// Check if the attachment network ID (state) matches the subnet network ID.
			attachmentNetworkID := data.NetworkID.ValueInt64()
			if subnetNetwork.ID != data.NetworkID.ValueInt64() {
				resp.Diagnostics.AddAttributeWarning(
					path.Root("subnet_id"),
					"Attachment network is different than the subnet network",
					fmt.Sprintf(
						"Attachment network (%d) is different that the subnet network (%d) (%s).",
						attachmentNetworkID, subnetNetwork.ID, data.SubnetID.ValueString(),
					),
				)

				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("network_id"), types.Int64Value(subnetNetwork.ID))...)
				resp.RequiresReplace.Append(path.Root("network_id"))
			}
		}

		if !data.IP.IsUnknown() && !data.IP.IsNull() {
			// Check if the attachment IP (state) is within the subnet ip range.
			attachmentIP := net.ParseIP(data.IP.ValueString())
			if !subnetIPRange.Contains(attachmentIP) {
				resp.Diagnostics.AddAttributeWarning(
					path.Root("subnet_id"),
					"Attachment IP is outside subnet IP range",
					fmt.Sprintf(
						"Attachment IP (%s) is outside subnet IP range (%s) (%s).",
						attachmentIP.String(), subnetIPRange, data.SubnetID.ValueString(),
					),
				)

				// Only marking the attribute as "RequiresReplace" does not work, as
				// terraform core internally filters out any replacements to attributes that
				// have not changed. Marking the attribute as unknown is correct, and makes
				// it so the RequiresReplace is actually applied and shown to the user.
				resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("ip"), iptypes.NewIPAddressUnknown())...)
				resp.RequiresReplace.Append(path.Root("ip"))
			}
		}
	}
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkResourceData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := &hcloud.Server{ID: data.ServerID.ValueInt64()}

	opts := hcloud.ServerAttachToNetworkOpts{}

	if !data.NetworkID.IsUnknown() && !data.NetworkID.IsNull() {
		opts.Network = &hcloud.Network{ID: data.NetworkID.ValueInt64()}
	}
	if !data.SubnetID.IsUnknown() && !data.SubnetID.IsNull() {
		subnetNetwork, subnetIPRange, err := r.ParseSubnetID(data.SubnetID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("subnet_id"),
				"Invalid Subnet ID",
				util.TitleCase(err.Error()),
			)
		}
		opts.Network = subnetNetwork
		opts.IPRange = subnetIPRange
	}

	if !data.IP.IsUnknown() && !data.IP.IsNull() {
		opts.IP = net.ParseIP(data.IP.ValueString())
	}

	if !data.AliasIPs.IsUnknown() && !data.AliasIPs.IsNull() {
		aliasIPsRaw := make([]string, 0, len(data.AliasIPs.Elements()))
		resp.Diagnostics.Append(data.AliasIPs.ElementsAs(ctx, &aliasIPsRaw, false)...)

		opts.AliasIPs = sliceutil.Transform(aliasIPsRaw, net.ParseIP)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Apply changes
	var action *hcloud.Action

	err := control.Retry(control.DefaultRetries, func() error {
		var innerErr error

		action, _, innerErr = r.client.Server.AttachToNetwork(ctx, server, opts)
		if hcloud.IsError(innerErr,
			hcloud.ErrorCodeConflict,
			hcloud.ErrorCodeLocked,
			hcloud.ErrorCodeServiceError,
			hcloud.ErrorCodeNoSubnetAvailable,
		) {
			return innerErr
		}
		if innerErr != nil {
			return control.AbortRetry(innerErr)
		}
		return nil
	})
	if err != nil {
		if !hcloud.IsError(err, hcloud.ErrorCodeServerAlreadyAttached) {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
	}
	if err := hcloudutil.WaitForAction(ctx, &r.client.Action, action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	// Refresh server
	server, _, err = r.client.Server.GetByID(ctx, server.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if server == nil {
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Server (%s) vanished", data.ServerID),
		)
		return
	}

	attachment := server.PrivateNetFor(opts.Network)
	if attachment == nil {
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Attachment of server (%s) to network (%s) vanished", data.ServerID, data.NetworkID),
		)
		return
	}

	// Populate data
	resp.Diagnostics.Append(populateNetworkResourceData(ctx, &data, server, attachment)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkResourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, network, err := r.ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	server, _, err = r.client.Server.GetByID(ctx, server.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if server == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	attachment := server.PrivateNetFor(network)
	if attachment == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(populateNetworkResourceData(ctx, &data, server, attachment)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, plan networkResourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, network, err := r.ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	if !plan.AliasIPs.Equal(data.AliasIPs) {
		opts := hcloud.ServerChangeAliasIPsOpts{
			Network: network,
		}

		aliasIPsRaw := make([]string, 0, len(plan.AliasIPs.Elements()))
		resp.Diagnostics.Append(plan.AliasIPs.ElementsAs(ctx, &aliasIPsRaw, false)...)

		opts.AliasIPs = sliceutil.Transform(aliasIPsRaw, net.ParseIP)

		// If data conversion failed we should abort before sending API requests.
		if resp.Diagnostics.HasError() {
			return
		}

		action, _, err := r.client.Server.ChangeAliasIPs(ctx, server, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if err := hcloudutil.WaitForAction(ctx, &r.client.Action, action); err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
	}

	server, _, err = r.client.Server.GetByID(ctx, server.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if server == nil {
		// Should not happen
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Server (%s) vanished", data.ServerID),
		)
		return
	}

	attachment := server.PrivateNetFor(network)
	if attachment == nil {
		// Should not happen
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Attachment of server (%s) to network (%s) vanished", data.ServerID, data.NetworkID),
		)
		return
	}

	resp.Diagnostics.Append(populateNetworkResourceData(ctx, &data, server, attachment)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkResourceData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, network, err := r.ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	opts := hcloud.ServerDetachFromNetworkOpts{
		Network: network,
	}

	var action *hcloud.Action
	err = control.Retry(control.DefaultRetries, func() error {
		var innerErr error

		action, _, innerErr = r.client.Server.DetachFromNetwork(ctx, server, opts)
		if hcloud.IsError(innerErr,
			hcloud.ErrorCodeConflict,
			hcloud.ErrorCodeLocked,
			hcloud.ErrorCodeServiceError) {
			return innerErr
		}
		return control.AbortRetry(innerErr)
	})
	if err != nil {
		if !hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
	}

	if err := hcloudutil.WaitForAction(ctx, &r.client.Action, action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NetworkResource) ParseID(s string) (*hcloud.Server, *hcloud.Network, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("unexpected id '%s', expected '$SERVER_ID-$NETWORK_ID'", s)
	}

	serverID, err := util.ParseID(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("unexpected id '%s', expected '$SERVER_ID-$NETWORK_ID'", s)
	}

	networkID, err := util.ParseID(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("unexpected id '%s', expected '$SERVER_ID-$NETWORK_ID'", s)
	}

	return &hcloud.Server{ID: serverID}, &hcloud.Network{ID: networkID}, nil
}

func (r *NetworkResource) ParseSubnetID(s string) (*hcloud.Network, *net.IPNet, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("unexpected subnet id '%s', expected '$NETWORK_ID-$SUBNET_IP_RANGE'", s)
	}

	networkID, err := util.ParseID(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("unexpected subnet id '%s', expected '$NETWORK_ID-$SUBNET_IP_RANGE'", s)
	}

	_, ipRange, err := net.ParseCIDR(parts[1])
	if ipRange == nil || err != nil {
		return nil, nil, fmt.Errorf("unexpected subnet id '%s', expected '$NETWORK_ID-$SUBNET_IP_RANGE'", s)
	}

	return &hcloud.Network{ID: networkID}, ipRange, nil
}
