package primaryip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const DataSourceListType = "hcloud_primary_ips"

var _ datasource.DataSource = (*DataSourceList)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSourceList)(nil)

type DataSourceList struct {
	client *hcloud.Client
}

func NewDataSourceList() datasource.DataSource {
	return &DataSourceList{}
}

func (d *DataSourceList) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceListType
}

func (d *DataSourceList) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *DataSourceList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a list of Hetzner Cloud Primary IPs.

See the [Primary IPs API documentation](https://docs.hetzner.cloud/reference/cloud#tag/primary-ips) for more details.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"primary_ips": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSourceSchema(true),
			},
			Computed: true,
		},
		"with_selector": schema.StringAttribute{
			MarkdownDescription: "Filter results using a [Label Selector](https://docs.hetzner.cloud/reference/cloud#label-selector)",
			Optional:            true,
		},
	}
}

type dataSourceListModel struct {
	ID         types.String `tfsdk:"id"`
	PrimaryIPs types.List   `tfsdk:"primary_ips"`

	WithSelector types.String `tfsdk:"with_selector"`
}

var _ util.ModelFromAPI[[]*hcloud.PrimaryIP] = &dataSourceListModel{}

func (m *dataSourceListModel) FromAPI(ctx context.Context, in []*hcloud.PrimaryIP) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	ids := make([]string, 0, len(in))
	tfItems := make([]attr.Value, 0, len(in))
	for _, item := range in {
		ids = append(ids, util.FormatID(item.ID))

		var value model
		diags.Append(value.FromAPI(ctx, item)...)

		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
	}

	m.ID = types.StringValue(datasourceutil.ListID(ids))
	m.PrimaryIPs, newDiags = types.ListValue((&model{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return diags
}

func (d *DataSourceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceListModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result []*hcloud.PrimaryIP
	var err error

	opts := hcloud.PrimaryIPListOpts{}
	if !data.WithSelector.IsNull() {
		opts.LabelSelector = data.WithSelector.ValueString()
	}

	result, err = d.client.PrimaryIP.AllWithOpts(ctx, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
