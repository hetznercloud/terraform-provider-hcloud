package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceServiceType is the type name of the Hetzner Cloud Load Balancer Service resource.
const DataSourceServiceType = "hcloud_load_balancer_service"

func getCommonServiceDataSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional: !readOnly,
			Computed: true,
		},
		"load_balancer_id": schema.Int64Attribute{
			Optional: !readOnly,
			Computed: true,
		},
		"protocol": schema.StringAttribute{
			Computed: true,
		},
		"listen_port": schema.Int32Attribute{
			Optional: !readOnly,
			Computed: true,
		},
		"destination_port": schema.Int32Attribute{
			Computed: true,
		},
		"proxyprotocol": schema.BoolAttribute{
			Computed: true,
		},
		"http": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"sticky_sessions": schema.BoolAttribute{
					Computed: true,
				},
				"cookie_name": schema.StringAttribute{
					Computed: true,
				},
				"cookie_lifetime": schema.Int32Attribute{
					Computed: true,
				},
				"certificates": schema.ListAttribute{
					Computed:    true,
					ElementType: types.Int64Type,
				},
				"redirect_http": schema.BoolAttribute{
					Computed: true,
				},
				"timeout_idle": schema.Int32Attribute{
					Computed: true,
				},
			},
		},
		"health_check": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"protocol": schema.StringAttribute{
					Computed: true,
				},
				"port": schema.Int32Attribute{
					Computed: true,
				},
				"interval": schema.Int32Attribute{
					Computed: true,
				},
				"timeout": schema.Int32Attribute{
					Computed: true,
				},
				"retries": schema.Int32Attribute{
					Computed: true,
				},
				"http": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Computed: true,
						},
						"path": schema.StringAttribute{
							Computed: true,
						},
						"response": schema.StringAttribute{
							Computed: true,
						},
						"tls": schema.BoolAttribute{
							Computed: true,
						},
						"status_codes": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
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
var _ datasource.DataSourceWithConfigValidators = (*DataSourceService)(nil)

type DataSourceService struct {
	client *hcloud.Client
}

func NewDataSourceService() datasource.DataSource {
	return &DataSourceService{}
}

// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
func (d *DataSourceService) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("load_balancer_id"),
			path.MatchRoot("listen_port"),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("load_balancer_id"),
			path.MatchRoot("listen_port"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("load_balancer_id"),
		),
	}
}

// Metadata should return the full name of the data source.
func (d *DataSourceService) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceServiceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSourceService) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema should return the schema for this data source.
func (d *DataSourceService) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Provides details about a Hetzner Cloud Load Balancer Service."
	resp.Schema.Attributes = getCommonServiceDataSchema(false)
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
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
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load_balancer", "id", data.LoadBalancerID.String()))
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
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load_balancer_service", "listen_port", data.ListenPort.String()))
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
