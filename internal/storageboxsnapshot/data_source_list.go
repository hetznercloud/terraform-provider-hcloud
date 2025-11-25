package storageboxsnapshot

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceListType is the type name of the Hetzner Storage Box Snapshots data source.
const DataSourceListType = "hcloud_storage_box_snapshots"

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
	experimental.StorageBox.AppendDiagnostic(&resp.Diagnostics)

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
Provides a list of Hetzner Storage Box Snapshots.

See the [Storage Box Snapshots API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots) for more details.
`

	experimental.StorageBox.AppendNotice(&resp.Schema.MarkdownDescription)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"storage_box_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box.",
			Required:            true,
		},
		"snapshots": schema.ListNestedAttribute{
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
	StorageBoxID types.Int64 `tfsdk:"storage_box_id"`
	Snapshots    types.List  `tfsdk:"snapshots"`

	WithSelector types.String `tfsdk:"with_selector"`
}

var _ util.ModelFromAPI[[]*hcloud.StorageBoxSnapshot] = &dataSourceListModel{}

func (m *dataSourceListModel) FromAPI(ctx context.Context, in []*hcloud.StorageBoxSnapshot) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	tfItems := make([]attr.Value, 0, len(in))

	for _, item := range in {
		var value dataSourceCommonModel
		diags.Append(value.FromAPI(ctx, item)...)

		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
	}

	m.Snapshots, newDiags = types.ListValue((&dataSourceCommonModel{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return diags
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *DataSourceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceListModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageBox := &hcloud.StorageBox{ID: data.StorageBoxID.ValueInt64()}

	var result []*hcloud.StorageBoxSnapshot
	var err error

	opts := hcloud.StorageBoxSnapshotListOpts{}
	if !data.WithSelector.IsNull() {
		opts.LabelSelector = data.WithSelector.ValueString()
	}

	result, err = d.client.StorageBox.AllSnapshotsWithOpts(ctx, storageBox, opts)
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
