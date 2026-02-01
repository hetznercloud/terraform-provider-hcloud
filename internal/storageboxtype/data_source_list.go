package storageboxtype

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceListType is the type name of the Hetzner Storage Box Types data source.
const DataSourceListType = "hcloud_storage_box_types"

var _ datasource.DataSource = (*DataSourceList)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSourceList)(nil)

type DataSourceList struct {
	client *hcloud.Client
}

func NewDataSourceList() datasource.DataSource {
	return &DataSourceList{}
}

// Metadata should return the full name of the data source.
func (d *DataSourceList) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceListType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSourceList) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema should return the schema for this data source.
func (d *DataSourceList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a list of Hetzner Storage Box Types.

See the [Storage Box Type API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-box-types) for more details.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"storage_box_types": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSourceSchema(true),
			},
			Computed: true,
		},
	}
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *DataSourceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result []*hcloud.StorageBoxType
	var err error

	result, err = d.client.StorageBoxType.AllWithOpts(ctx, hcloud.StorageBoxTypeListOpts{})
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
