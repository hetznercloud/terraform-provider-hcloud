package loadbalancer

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/validateutil"
)

const NetworkResourceType = "hcloud_load_balancer_network"

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
Manage the attachment of a Load Balancer in a Network in the Hetzner Cloud.
`)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"load_balancer_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Load Balancer.",
			Required:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"network_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Network to attach the Load Balancer to. Using `subnet_id` is preferred. Required if `subnet_id` is not set. If `subnet_id` or `ip` are not set, the Load Balancer will be attached to the last subnet (ordered by `ip_range`).",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"subnet_id": schema.StringAttribute{
			MarkdownDescription: "ID of the Subnet to attach the Load Balancer to. Required if `network_id` is not set.",
			Optional:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"ip": schema.StringAttribute{
			CustomType:          iptypes.IPAddressType{},
			MarkdownDescription: "IP to assign to the Load Balancer.",
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
		// XXX: Move to `load_balancer` since it is unrelated to the given private
		// network attachment.
		"enable_public_interface": schema.BoolAttribute{
			MarkdownDescription: "Wether the Load Balancer public interface is enabled. Default is `true`.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
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

	loadBalancer := &hcloud.LoadBalancer{ID: data.LoadBalancerID.ValueInt64()}

	opts := hcloud.LoadBalancerAttachToNetworkOpts{}

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

	if resp.Diagnostics.HasError() {
		return
	}

	// Apply changes
	var action *hcloud.Action

	err := control.Retry(control.DefaultRetries, func() error {
		var innerErr error

		action, _, innerErr = r.client.LoadBalancer.AttachToNetwork(ctx, loadBalancer, opts)
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
		// Do not fail if the load balancer is already attached.
		if !hcloud.IsError(err, hcloud.ErrorCodeLoadBalancerAlreadyAttached) {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		// A load balancer cannot be attached to more than one network, fail if the
		// attached network is not the one we requested.
		{
			lb, _, innerErr := r.client.LoadBalancer.GetByID(ctx, loadBalancer.ID)
			if innerErr != nil {
				resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(innerErr)...)
				return
			}
			if lb.PrivateNetFor(opts.Network) == nil {
				resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
				return
			}
		}
	}
	if err := r.client.Action.WaitFor(ctx, action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
	// Make sure to save the ID immediately so we can recover if the process stops after
	// this call. Terraform marks the resource as "tainted", so it can be deleted and no
	// surprise "duplicate resource" errors happen.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(fmt.Sprintf("%d-%d", loadBalancer.ID, opts.Network.ID)))...)

	// Refresh
	loadBalancer, _, err = r.client.LoadBalancer.GetByID(ctx, loadBalancer.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if loadBalancer == nil {
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Load Balancer (%s) vanished", data.LoadBalancerID),
		)
		return
	}

	// XXX: Toggle public interface
	{
		resp.Diagnostics.Append(r.setLoadBalancerPublicInterfaceEnabled(ctx, loadBalancer, data.EnablePublicInterface.ValueBool())...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	attachment := loadBalancer.PrivateNetFor(opts.Network)
	if attachment == nil {
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Attachment of load balancer (%s) to network (%s) vanished", data.LoadBalancerID, data.NetworkID),
		)
		return
	}

	// Populate data
	resp.Diagnostics.Append(populateNetworkResourceData(ctx, &data, loadBalancer, attachment)...)
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

	loadBalancer, network, err := r.ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	loadBalancer, _, err = r.client.LoadBalancer.GetByID(ctx, loadBalancer.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if loadBalancer == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	attachment := loadBalancer.PrivateNetFor(network)
	if attachment == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(populateNetworkResourceData(ctx, &data, loadBalancer, attachment)...)
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

	loadBalancer, network, err := r.ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	loadBalancer, _, err = r.client.LoadBalancer.GetByID(ctx, loadBalancer.ID)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	if loadBalancer == nil {
		// Should not happen
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Load Balancer (%s) vanished", data.LoadBalancerID),
		)
		return
	}

	// XXX: Toggle public interface
	{
		if !plan.EnablePublicInterface.Equal(data.EnablePublicInterface) {
			resp.Diagnostics.Append(r.setLoadBalancerPublicInterfaceEnabled(ctx, loadBalancer, plan.EnablePublicInterface.ValueBool())...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	attachment := loadBalancer.PrivateNetFor(network)
	if attachment == nil {
		// Should not happen
		resp.Diagnostics.AddError(
			"Resource vanished",
			fmt.Sprintf("Attachment of load balancer (%s) to network (%s) vanished", data.LoadBalancerID, data.NetworkID),
		)
		return
	}

	resp.Diagnostics.Append(populateNetworkResourceData(ctx, &data, loadBalancer, attachment)...)
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

	loadBalancer, network, err := r.ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	opts := hcloud.LoadBalancerDetachFromNetworkOpts{
		Network: network,
	}

	var action *hcloud.Action
	err = control.Retry(control.DefaultRetries, func() error {
		var innerErr error

		action, _, innerErr = r.client.LoadBalancer.DetachFromNetwork(ctx, loadBalancer, opts)
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

	if err := r.client.Action.WaitFor(ctx, action); err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	{
		// Workaround some flakiness from the API
		for range 10 {
			lb, _, err := r.client.LoadBalancer.GetByID(ctx, loadBalancer.ID)
			if err == nil {
				if lb == nil || lb.PrivateNetFor(opts.Network) == nil {
					break
				}
			}

			time.Sleep(2 * time.Second)
		}
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NetworkResource) ParseID(s string) (*hcloud.LoadBalancer, *hcloud.Network, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("unexpected id '%s', expected '$LOAD_BALANCER_ID-$NETWORK_ID'", s)
	}

	loadBalancerID, err := util.ParseID(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("unexpected id '%s', expected '$LOAD_BALANCER_ID-$NETWORK_ID'", s)
	}

	networkID, err := util.ParseID(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("unexpected id '%s', expected '$LOAD_BALANCER_ID-$NETWORK_ID'", s)
	}

	return &hcloud.LoadBalancer{ID: loadBalancerID}, &hcloud.Network{ID: networkID}, nil
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

func (r *NetworkResource) setLoadBalancerPublicInterfaceEnabled(ctx context.Context, loadBalancer *hcloud.LoadBalancer, enabled bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if loadBalancer.PublicNet.Enabled == enabled {
		return diags
	}

	var action *hcloud.Action
	var err error
	if enabled {
		action, _, err = r.client.LoadBalancer.EnablePublicInterface(ctx, loadBalancer)
	} else {
		action, _, err = r.client.LoadBalancer.DisablePublicInterface(ctx, loadBalancer)
	}

	if err != nil {
		diags.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return diags
	}
	if err = r.client.Action.WaitFor(ctx, action); err != nil {
		diags.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return diags
	}

	loadBalancer.PublicNet.Enabled = enabled

	return diags
}
