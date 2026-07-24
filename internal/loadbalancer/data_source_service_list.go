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

const DataSourceServiceListType = "hcloud_load_balancer_services"

var _ datasource.DataSource = (*DataSourceServiceList)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSourceServiceList)(nil)

type DataSourceServiceList struct {
	client *hcloud.Client
}

func NewDataSourceServiceList() datasource.DataSource {
	return &DataSourceServiceList{}
}

func (d *DataSourceServiceList) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceServiceListType
}

func (d *DataSourceServiceList) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *DataSourceServiceList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Provides a list of Hetzner Cloud Load Balancer Services."

	resp.Schema.Attributes = map[string]schema.Attribute{
		"load_balancer_id": schema.Int64Attribute{
			MarkdownDescription: "ID of Load Balancer to fetch services from.",
			Required:            true,
		},
		"services": schema.ListNestedAttribute{
			MarkdownDescription: "List of the Load Balancer's services.",
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
		resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load balancer", "id", data.LoadBalancerID.String()))
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
