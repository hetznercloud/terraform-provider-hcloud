package rdns

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/validateutil"
)

const ResourceType = "hcloud_rdns"

var _ resource.Resource = (*Resource)(nil)
var _ resource.ResourceWithConfigure = (*Resource)(nil)
var _ resource.ResourceWithImportState = (*Resource)(nil)
var _ resource.ResourceWithConfigValidators = (*Resource)(nil)

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
	resp.Schema.MarkdownDescription = `
Provides Hetzner Cloud reverse DNS (rDNS) entries for Servers, Primary IPs, Floating IPs or Load Balancers.
`
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Reverse DNS entry. Formatted like `$RESOURCE_PREFIX-$RESOURCE_ID-$IP_ADDRESS`, where `$RESOURCE_PREFIX` is `s` for Servers, `p` for Primary IPs, `f` for Floating IPs and `l` for Load Balancers.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"server_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Server the `ip_address` belongs to.",
			Optional:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"primary_ip_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Primary IP the `ip_address` belongs to.",
			Optional:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"floating_ip_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Floating IP the `ip_address` belongs to.",
			Optional:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"load_balancer_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Load Balancer the `ip_address` belongs to.",
			Optional:            true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"ip_address": schema.StringAttribute{
			CustomType:          iptypes.IPAddressType{},
			MarkdownDescription: "IP address that should point to `dns_ptr`.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				validateutil.IP(),
			},
		},
		"dns_ptr": schema.StringAttribute{
			MarkdownDescription: "Domain name `ip_address` should point to.",
			Required:            true,
		},
	}
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("server_id"),
			path.MatchRoot("primary_ip_id"),
			path.MatchRoot("floating_ip_id"),
			path.MatchRoot("load_balancer_id"),
		),
	}
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ip := net.ParseIP(data.IPAddress.ValueString())
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("ip_address"),
			"Invalid Attribute Value",
			fmt.Sprintf("Attribute ip_address must be a valid ip, got: %s", data.IPAddress.ValueString()),
		)
	}

	dnsPtr := data.DNSPtr.ValueString()

	var rdns hcloud.RDNSSupporter

	switch {
	case !data.ServerID.IsUnknown() && !data.ServerID.IsNull():
		res, _, err := r.client.Server.GetByID(ctx, data.ServerID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("server", "id", data.ServerID.ValueInt64()))
			return
		}
		rdns = res

	case !data.PrimaryIPID.IsUnknown() && !data.PrimaryIPID.IsNull():
		res, _, err := r.client.PrimaryIP.GetByID(ctx, data.PrimaryIPID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("primary ip", "id", data.PrimaryIPID.ValueInt64()))
			return
		}
		rdns = res

	case !data.FloatingIPID.IsUnknown() && !data.FloatingIPID.IsNull():
		res, _, err := r.client.FloatingIP.GetByID(ctx, data.FloatingIPID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("floating ip", "id", data.FloatingIPID.ValueInt64()))
			return
		}
		rdns = res

	case !data.LoadBalancerID.IsUnknown() && !data.LoadBalancerID.IsNull():
		res, _, err := r.client.LoadBalancer.GetByID(ctx, data.LoadBalancerID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load balancer", "id", data.LoadBalancerID.ValueInt64()))
			return
		}
		rdns = res
	}

	if resp.Diagnostics.HasError() {
		return
	}

	action, _, err := r.client.RDNS.ChangeDNSPtr(ctx, rdns, ip, &dnsPtr)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, rdns, ip, dnsPtr)...)
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

	rdns, ip, err := ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	switch v := rdns.(type) {
	case *hcloud.Server:
		res, _, err := r.client.Server.GetByID(ctx, v.ID)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.State.RemoveResource(ctx)
			return
		}
		rdns = res

	case *hcloud.PrimaryIP:
		res, _, err := r.client.PrimaryIP.GetByID(ctx, v.ID)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.State.RemoveResource(ctx)
			return
		}
		rdns = res

	case *hcloud.FloatingIP:
		res, _, err := r.client.FloatingIP.GetByID(ctx, v.ID)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.State.RemoveResource(ctx)
			return
		}
		rdns = res

	case *hcloud.LoadBalancer:
		res, _, err := r.client.LoadBalancer.GetByID(ctx, v.ID)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if res == nil {
			resp.State.RemoveResource(ctx)
			return
		}
		rdns = res
	}

	dnsPtr, err := rdns.GetDNSPtrForIP(ip)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Write data to state
	resp.Diagnostics.Append(data.FromAPI(ctx, rdns, ip, dnsPtr)...)
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

	rdns, ip, err := ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	if !data.DNSPtr.Equal(plan.DNSPtr) {
		action, _, err := r.client.RDNS.ChangeDNSPtr(ctx, rdns, ip, plan.DNSPtr.ValueStringPointer())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Write data to state
	resp.Diagnostics.Append(data.FromAPI(ctx, rdns, ip, plan.DNSPtr.ValueString())...)
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

	rdns, ip, err := ParseID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ID",
			util.TitleCase(err.Error()),
		)
		return
	}

	action, _, err := r.client.RDNS.ChangeDNSPtr(ctx, rdns, ip, nil)
	if err != nil {
		if hcloudutil.APIErrorIsNotFound(err) {
			return
		}

		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &r.client.Action, action)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
