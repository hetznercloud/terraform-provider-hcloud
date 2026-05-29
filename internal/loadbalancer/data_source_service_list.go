package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceServiceListType is the type name to receive a list of Hetzner Cloud Load Balancer Service resources.
const DataSourceServiceListType = "hcloud_load_balancer_services"

var _ datasource.DataSource = (*DataSourceServiceList)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSourceServiceList)(nil)

type DataSourceServiceList struct {
	client *hcloud.Client
}

func NewDataSourceServiceList() datasource.DataSource {
	return &DataSourceServiceList{}
}

// Metadata should return the full name of the data source.
func (d *DataSourceServiceList) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceServiceListType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSourceServiceList) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema should return the schema for this data source.
func (d *DataSourceServiceList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Provides a list of Hetzner Cloud Load Balancer Services."

	resp.Schema.Attributes = map[string]schema.Attribute{
		"load_balancer_id": schema.Int64Attribute{
			Required: true,
		},
		"services": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonServiceDataSchema(true),
			},
			Computed: true,
		},
	}
}

type dataSourceServiceListModel struct {
	LoadBalancerID types.Int64 `tfsdk:"load_balancer_id"`
	Services       types.List  `tfsdk:"services"`
}

func populateDataSourceServiceListModel(ctx context.Context, data *dataSourceServiceListModel, lb *hcloud.LoadBalancer, services []*hcloud.LoadBalancerService) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	tfItems := make([]attr.Value, 0, len(services))

	for _, item := range services {
		var value serviceModel
		diags.Append(value.FromAPI(ctx, item)...)

		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
	}

	data.LoadBalancerID = types.Int64Value(lb.ID)
	data.Services, newDiags = types.ListValue((&serviceModel{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return diags
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *DataSourceServiceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceServiceListModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, _, err := d.client.LoadBalancer.GetByID(ctx, data.LoadBalancerID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
	if lb == nil {
		resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load_balancer", "id", data.LoadBalancerID.String()))
		return
	}

	var services []*hcloud.LoadBalancerService
	for _, svc := range lb.Services {
		services = append(services, &svc)
	}

	resp.Diagnostics.Append(populateDataSourceServiceListModel(ctx, &data, lb, services)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
