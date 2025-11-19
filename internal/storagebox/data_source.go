package storagebox

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceType is the type name of the Hetzner Storage Box data source.
const DataSourceType = "hcloud_storage_box"

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "Primary username of the Storage Box.",
			Computed:            true,
		},
		"storage_box_type": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box Type.",
			Computed:            true,
		},
		"location": schema.StringAttribute{
			MarkdownDescription: "Name of the Location.",
			Computed:            true,
		},
		"labels": schema.MapAttribute{
			MarkdownDescription: "User-defined [labels](https://docs.hetzner.cloud/reference/cloud#labels) (key-value pairs) for the resource.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"access_settings": schema.SingleNestedAttribute{
			MarkdownDescription: "Access settings of the Storage Box.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"reachable_externally": schema.BoolAttribute{
					MarkdownDescription: "Whether access from outside the Hetzner network is allowed.",
					Computed:            true,
				},
				"samba_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the Samba subsystem is enabled.",
					Computed:            true,
				},
				"ssh_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the SSH subsystem is enabled.",
					Computed:            true,
				},
				"webdav_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the WebDAV subsystem is enabled.",
					Computed:            true,
				},
				"zfs_enabled": schema.BoolAttribute{
					MarkdownDescription: "Whether the ZFS snapshot folder is visible.",
					Computed:            true,
				},
			},
		},
		"server": schema.StringAttribute{
			MarkdownDescription: "FQDN of the Storage Box.",
			Computed:            true,
		},
		"system": schema.StringAttribute{
			MarkdownDescription: "Host system of the Storage Box.",
			Computed:            true,
		},
		"delete_protection": schema.BoolAttribute{
			MarkdownDescription: "Whether delete protection is enabled.",
			Computed:            true,
		},
		"snapshot_plan": schema.SingleNestedAttribute{
			MarkdownDescription: "Details of the active snapshot plan.",
			Computed:            true,

			Attributes: map[string]schema.Attribute{
				"max_snapshots": schema.Int32Attribute{
					MarkdownDescription: "Maximum amount of Snapshots that will be created by this Snapshot Plan. Older Snapshots will be deleted.",
					Computed:            true,
				},
				"minute": schema.Int32Attribute{
					MarkdownDescription: "Minute when the Snapshot Plan is executed (UTC).",
					Computed:            true,
				},
				"hour": schema.Int32Attribute{
					MarkdownDescription: "Hour when the Snapshot Plan is executed (UTC).",
					Computed:            true,
				},
				"day_of_week": schema.Int32Attribute{
					MarkdownDescription: "Day of the week when the Snapshot Plan is executed. Starts at 0 for Sunday til 6 for Saturday. Note that this differs from the API, which uses 1 (Monday) through 7 (Sunday). Null means every day.",
					Computed:            true,
				},
				"day_of_month": schema.Int32Attribute{
					MarkdownDescription: "Day of the month when the Snapshot Plan is executed. Null means every day.",
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
Provides details about a Hetzner Storage Box.

See the [Storage Boxes API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-boxes) for more details.
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

type dataSourceModel struct {
	commonModel

	WithSelector types.String `tfsdk:"with_selector"`
}

var _ util.ModelFromAPI[*hcloud.StorageBox] = &dataSourceModel{}

func (m *dataSourceModel) FromAPI(ctx context.Context, in *hcloud.StorageBox) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(m.commonModel.FromAPI(ctx, in)...)

	return diags
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

	var result *hcloud.StorageBox
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.StorageBox.GetByID(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.StorageBox.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box", "name", data.Name.String()))
			return
		}
	case !data.WithSelector.IsNull():
		opts := hcloud.StorageBoxListOpts{}
		opts.LabelSelector = data.WithSelector.ValueString()

		all, err := d.client.StorageBox.AllWithOpts(ctx, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		var newDiag diag.Diagnostic
		result, newDiag = datasourceutil.GetOneResultForLabelSelector("storage box", all, opts.LabelSelector)
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
