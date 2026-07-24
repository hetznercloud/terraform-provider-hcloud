package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const DataSourceServiceType = "hcloud_load_balancer_service"

func getCommonServiceDataSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Load Balancer Service. Format: `<load_balancer_id>__<listen_port>`",
			Computed:            true,
		},
		"load_balancer_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Load Balancer this Service belongs to.",
			Required:            !readOnly,
			Computed:            readOnly,
		},
		"listen_port": schema.Int32Attribute{
			MarkdownDescription: "Port the Load Balancer listens on.",
			Required:            !readOnly,
			Computed:            readOnly,
		},
		"protocol": schema.StringAttribute{
			MarkdownDescription: "Protocol of the Load Balancer. One of `tcp`, `http`, `https`.",
			Computed:            true,
		},
		"destination_port": schema.Int32Attribute{
			MarkdownDescription: "Port the Load Balancer will balance to.",
			Computed:            true,
		},
		"proxyprotocol": schema.BoolAttribute{
			MarkdownDescription: "Whether the PROXY protocol is enabled.",
			Computed:            true,
		},
		"http": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for http(s) protocol.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"sticky_sessions": schema.BoolAttribute{
					MarkdownDescription: "Whether to use sticky sessions.",
					Computed:            true,
				},
				"cookie_name": schema.StringAttribute{
					MarkdownDescription: "Name of the cookie used for sticky sessions.",
					Computed:            true,
				},
				"cookie_lifetime": schema.Int32Attribute{
					MarkdownDescription: "Lifetime of the cookie used for sticky sessions (in seconds).",
					Computed:            true,
				},
				"certificate_ids": schema.ListAttribute{
					MarkdownDescription: "IDs of the Certificates to use for TLS/SSL termination by the Load Balancer; empty for TLS/SSL passthrough.",
					Computed:            true,
					ElementType:         types.Int64Type,
				},
				"redirect_http": schema.BoolAttribute{
					MarkdownDescription: "Redirect HTTP requests to HTTPS.",
					Computed:            true,
				},
				"timeout_idle": schema.Int32Attribute{
					MarkdownDescription: "Idle timeout in seconds for the client and server side.",
					Computed:            true,
				},
			},
		},
		"health_check": schema.SingleNestedAttribute{
			MarkdownDescription: "Service health check.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"protocol": schema.StringAttribute{
					MarkdownDescription: "Type of the health check. One of `tcp`, `http`.",
					Computed:            true,
				},
				"port": schema.Int32Attribute{
					MarkdownDescription: "Port the health check will be performed on.",
					Computed:            true,
				},
				"interval": schema.Int32Attribute{
					MarkdownDescription: "Time interval in seconds health checks are performed.",
					Computed:            true,
				},
				"timeout": schema.Int32Attribute{
					MarkdownDescription: "Time in seconds after an attempt is considered a timeout.",
					Computed:            true,
				},
				"retries": schema.Int32Attribute{
					MarkdownDescription: "Unsuccessful retries needed until a target is considered unhealthy; an unhealthy target needs the same number of successful retries to become healthy again.",
					Computed:            true,
				},
				"http": schema.SingleNestedAttribute{
					MarkdownDescription: "Additional configuration for protocol http.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							MarkdownDescription: "Host header to send in the HTTP request.",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "HTTP path to use for health checks.",
							Computed:            true,
						},
						"response": schema.StringAttribute{
							MarkdownDescription: "String that must be contained in HTTP response in order to pass the health check.",
							Computed:            true,
						},
						"tls": schema.BoolAttribute{
							MarkdownDescription: "Use HTTPS for health check.",
							Computed:            true,
						},
						"status_codes": schema.ListAttribute{
							MarkdownDescription: "List of returned HTTP status codes in order to pass the health check. Supports the wildcards ? for exactly one character and * for multiple ones.",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

type dataSourceServiceModel struct {
	serviceModel
}

func populateDataSourceServiceModel(ctx context.Context, data *dataSourceServiceModel, lb *hcloud.LoadBalancer, svc *hcloud.LoadBalancerService) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(data.FromAPI(ctx, svc)...)
	data.ID = types.StringValue(fmt.Sprintf("%d__%d", lb.ID, svc.ListenPort))
	data.LoadBalancerID = types.Int64Value(lb.ID)

	return diags
}

var _ datasource.DataSource = (*DataSourceService)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSourceService)(nil)

type DataSourceService struct {
	client *hcloud.Client
}

func NewDataSourceService() datasource.DataSource {
	return &DataSourceService{}
}

func (d *DataSourceService) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceServiceType
}

func (d *DataSourceService) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *DataSourceService) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Provides details about a Hetzner Cloud Load Balancer Service."
	resp.Schema.Attributes = getCommonServiceDataSchema(false)
}

func (d *DataSourceService) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceServiceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		lb  *hcloud.LoadBalancer
		svc *hcloud.LoadBalancerService
		err error
	)

	switch {
	case !data.ID.IsNull():
		lb, svc, err = lookupLoadBalancerServiceID(ctx, data.ID.ValueString(), d.client)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

	case !data.LoadBalancerID.IsNull() && !data.ListenPort.IsNull():
		lb, _, err = d.client.LoadBalancer.GetByID(ctx, data.LoadBalancerID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if lb == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load balancer", "id", data.LoadBalancerID.String()))
			return
		}

		listenPort := int(data.ListenPort.ValueInt32())
		for _, _svc := range lb.Services {
			if _svc.ListenPort == listenPort {
				svc = &_svc
				break
			}
		}
		if svc == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load balancer service", "listen_port", data.ListenPort.String()))
			return
		}
	}

	if lb == nil || svc == nil {
		return
	}

	resp.Diagnostics.Append(populateDataSourceServiceModel(ctx, &data, lb, svc)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
