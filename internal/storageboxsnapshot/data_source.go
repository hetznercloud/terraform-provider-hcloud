package storageboxsnapshot

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/merge"
)

// DataSourceType is the type name of the Hetzner Storage Box Snapshot data source.
const DataSourceType = "hcloud_storage_box_snapshot"

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"storage_box_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box.",
			Required:            !readOnly,
			Computed:            readOnly,
		},
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box Snapshot.",
			Optional:            !readOnly,
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box Snapshot.",
			Optional:            !readOnly,
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Storage Box Snapshot.",
			Computed:            true,
		},
		"is_automatic": schema.BoolAttribute{
			MarkdownDescription: "Whether the Storage Box Snapshot was created automatically.",
			Computed:            true,
		},
		"labels": datasourceutil.LabelsSchema(),
		"stats": schema.SingleNestedAttribute{
			MarkdownDescription: "Statistics of the Storage Box Snapshot.",
			Computed:            true,

			Attributes: map[string]schema.Attribute{
				"size": schema.Int64Attribute{
					MarkdownDescription: "Current storage requirements of the Snapshot in bytes.",
					Computed:            true,
				},
				"size_filesystem": schema.Int64Attribute{
					MarkdownDescription: "Size of the compressed file system contained in the Snapshot in bytes.\n",
					Computed:            true,
				},
			},
		},
	}
}

// Single
var _ datasource.DataSource = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigValidators = (*DataSource)(nil)

type DataSource struct {
	client *hcloud.Client
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata should return the full name of the data source.
func (d *DataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	experimental.StorageBox.AppendDiagnostic(&resp.Diagnostics)

	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Schema should return the schema for this data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides details about a Hetzner Storage Box Snapshot.

See the [Storage Box Snapshots API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-box-snapshots) for more details.
`

	experimental.StorageBox.AppendNotice(&resp.Schema.MarkdownDescription)

	resp.Schema.Attributes = getCommonDataSourceSchema(false)
	maps.Copy(resp.Schema.Attributes, map[string]schema.Attribute{
		"with_selector": schema.StringAttribute{
			MarkdownDescription: "Filter results using a [Label Selector](https://docs.hetzner.cloud/reference/hetzner#label-selector).",
			Optional:            true,
		},
	})
}

type dataSourceCommonModel struct {
	model

	Stats types.Object `tfsdk:"stats"`
}

var _ util.ModelFromAPI[*hcloud.StorageBoxSnapshot] = &dataSourceCommonModel{}
var _ util.ModelToTerraform[types.Object] = &dataSourceCommonModel{}

func (m *dataSourceCommonModel) tfAttributesTypes() map[string]attr.Type {
	return merge.Maps(
		m.model.tfAttributesTypes(),
		map[string]attr.Type{
			"stats": types.ObjectType{AttrTypes: (&modelStats{}).tfAttributesTypes()},
		})
}

func (m *dataSourceCommonModel) tfType() attr.Type {
	return basetypes.ObjectType{AttrTypes: m.tfAttributesTypes()}
}

func (m *dataSourceCommonModel) ToTerraform(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, m.tfAttributesTypes(), m)
}

func (m *dataSourceCommonModel) FromAPI(ctx context.Context, hc *hcloud.StorageBoxSnapshot) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	diags.Append(m.model.FromAPI(ctx, hc)...)

	{
		value := modelStats{}
		diags.Append(value.FromAPI(ctx, hc.Stats)...)

		m.Stats, newDiags = value.ToTerraform(ctx)
		diags.Append(newDiags...)
	}

	return diags
}

type dataSourceModel struct {
	dataSourceCommonModel

	WithSelector types.String `tfsdk:"with_selector"`
}

// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
func (d *DataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
			path.MatchRoot("with_selector"),
		),
	}
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageBox := &hcloud.StorageBox{
		ID: data.StorageBoxID.ValueInt64(),
	}

	var result *hcloud.StorageBoxSnapshot
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.StorageBox.GetSnapshotByID(ctx, storageBox, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box snapshot", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.StorageBox.GetSnapshotByName(ctx, storageBox, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box snapshot", "name", data.Name.String()))
			return
		}
	case !data.WithSelector.IsNull():
		opts := hcloud.StorageBoxSnapshotListOpts{}
		opts.LabelSelector = data.WithSelector.ValueString()

		all, err := d.client.StorageBox.AllSnapshotsWithOpts(ctx, storageBox, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		var newDiag diag.Diagnostic
		result, newDiag = datasourceutil.GetOneResultForLabelSelector("storage box snapshot", all, opts.LabelSelector)
		if newDiag != nil {
			resp.Diagnostics.Append(newDiag)
			return
		}
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
