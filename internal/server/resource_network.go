package server

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/sliceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const NetworkResourceType = "hcloud_server_network"

var _ resource.Resource = (*networkResourceImpl)(nil)
var _ resource.ResourceWithConfigure = (*networkResourceImpl)(nil)
var _ resource.ResourceWithConfigValidators = (*networkResourceImpl)(nil)
var _ resource.ResourceWithImportState = (*networkResourceImpl)(nil)

type networkResourceImpl struct {
	client *hcloud.Client
}

func NewNetworkResource() resource.Resource {
	return &networkResourceImpl{}
}

// Metadata should return the full name of the data source.
func (r *networkResourceImpl) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = NetworkResourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (r *networkResourceImpl) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	r.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *networkResourceImpl) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = util.MarkdownDescription(`
		Manage the attachment of a Server in a Network in the Hetzner Cloud.

		If ''subnet_id'' or ''ip'' are not provided, the Server will be assigned an IP in the last created subnet.
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
			MarkdownDescription: "ID of the Network to attach the Server to. Using `subnet_id` is preferred. Required if `subnet_id` is not set.",
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
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"ip": schema.StringAttribute{
			MarkdownDescription: "IP to assign to the Server.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"alias_ips": schema.SetAttribute{
			MarkdownDescription: "Additional IPs to assign to the Server.",
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
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

// ConfigValidators returns a list of functions which will all be performed during validation.
func (r *networkResourceImpl) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("network_id"),
			path.MatchRoot("subnet_id"),
		),
	}
}

type networkResourceData struct {
	ID         types.String `tfsdk:"id"`
	ServerID   types.Int64  `tfsdk:"server_id"`
	NetworkID  types.Int64  `tfsdk:"network_id"`
	SubnetID   types.String `tfsdk:"subnet_id"`
	IP         types.String `tfsdk:"ip"`
	AliasIPs   types.Set    `tfsdk:"alias_ips"`
	MACAddress types.String `tfsdk:"mac_address"`
}

func populateNetworkResourceData(
	ctx context.Context,
	data *networkResourceData,
	server *hcloud.Server,
	attachment *hcloud.ServerPrivateNet,
) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	data.ID = types.StringValue(fmt.Sprintf("%d-%d", server.ID, attachment.Network.ID))
	data.ServerID = types.Int64Value(server.ID)
	data.NetworkID = types.Int64Value(attachment.Network.ID)
	data.IP = types.StringValue(attachment.IP.String())
	data.MACAddress = types.StringValue(attachment.MACAddress)

	if !data.AliasIPs.IsNull() || len(attachment.Aliases) > 0 {
		aliasIPsStrings := sliceutil.Transform(attachment.Aliases, func(e net.IP) string { return e.String() })

		data.AliasIPs, newDiags = types.SetValueFrom(ctx, types.StringType, aliasIPsStrings)
		diags.Append(newDiags...)
	}

	return diags
}

func (r *networkResourceImpl) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		} else {
			// Invalid case when a given network ID does not match the given subnet ID
			if opts.Network != nil && opts.Network.ID != subnetNetwork.ID {
				resp.Diagnostics.AddAttributeError(
					path.Root("subnet_id"),
					"Invalid Subnet ID",
					fmt.Sprintf(
						"The network ID (%s) and the subnet ID (%s) do not refer to the same network ID.",
						data.NetworkID,
						data.SubnetID,
					),
				)
			}
			opts.Network = subnetNetwork
			opts.IPRange = subnetIPRange
		}
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

func (r *networkResourceImpl) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	if !data.SubnetID.IsUnknown() && !data.SubnetID.IsNull() {
		_, subnetIPRange, err := r.ParseSubnetID(data.SubnetID.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("subnet_id"),
				"Invalid Subnet ID",
				util.TitleCase(err.Error()),
			)
		}
		if !subnetIPRange.Contains(attachment.IP) {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("subnet_id"),
				"Assigned IP is outside subnet IP range",
				"",
			)
		}
	}

	resp.Diagnostics.Append(populateNetworkResourceData(ctx, &data, server, attachment)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *networkResourceImpl) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	if !plan.AliasIPs.IsUnknown() && !plan.AliasIPs.Equal(data.AliasIPs) {
		opts := hcloud.ServerChangeAliasIPsOpts{
			Network: network,
		}

		aliasIPsRaw := make([]string, 0, len(plan.AliasIPs.Elements()))
		resp.Diagnostics.Append(plan.AliasIPs.ElementsAs(ctx, &aliasIPsRaw, false)...)

		opts.AliasIPs = sliceutil.Transform(aliasIPsRaw, net.ParseIP)

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

func (r *networkResourceImpl) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *networkResourceImpl) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *networkResourceImpl) ParseID(s string) (*hcloud.Server, *hcloud.Network, error) {
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

func (r *networkResourceImpl) ParseSubnetID(s string) (*hcloud.Network, *net.IPNet, error) {
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
