package storageboxsubaccount

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceType is the type name of the Hetzner Storage Box Subaccount data source.
const DataSourceType = "hcloud_storage_box_subaccount"

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"storage_box_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box.",
			Required:            !readOnly,
			Computed:            readOnly,
		},
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Storage Box Subaccount.",
			Optional:            !readOnly,
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Storage Box Subaccount.",
			Optional:            !readOnly,
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Storage Box Subaccount.",
			Computed:            true,
		},
		"home_directory": schema.StringAttribute{
			MarkdownDescription: "Home directory of the Storage Box Subaccount.",
			Computed:            true,
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "Username of the Storage Box Subaccount.",
			Optional:            !readOnly,
			Computed:            true,
		},
		"server": schema.StringAttribute{
			MarkdownDescription: "FQDN of the Storage Box Subaccount.",
			Computed:            true,
		},
		"access_settings": schema.SingleNestedAttribute{
			MarkdownDescription: "Access settings for the Subaccount.",
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
				"readonly": schema.BoolAttribute{
					MarkdownDescription: "Whether the Subaccount is read-only.",
					Computed:            true,
				},
			},
		},
		"labels": datasourceutil.LabelsSchema(),
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
Provides details about a Hetzner Storage Box Subaccount.

See the [Storage Box Subaccounts API documentation](https://docs.hetzner.cloud/reference/hetzner#storage-box-subaccounts) for more details.
`

	resp.Schema.Attributes = getCommonDataSourceSchema(false)
	maps.Copy(resp.Schema.Attributes, map[string]schema.Attribute{
		"with_selector": schema.StringAttribute{
			MarkdownDescription: "Filter results using a [Label Selector](https://docs.hetzner.cloud/reference/hetzner#label-selector).",
			Optional:            true,
		},
	})
}

type dataSourceModel struct {
	model
	WithSelector types.String `tfsdk:"with_selector"`
}

// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
func (d *DataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
			path.MatchRoot("username"),
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

	var result *hcloud.StorageBoxSubaccount
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.StorageBox.GetSubaccountByID(ctx, storageBox, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box subaccount", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.StorageBox.GetSubaccountByName(ctx, storageBox, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box subaccount", "name", data.Name.String()))
			return
		}
	case !data.Username.IsNull():
		result, _, err = d.client.StorageBox.GetSubaccountByUsername(ctx, storageBox, data.Username.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("storage box subaccount", "username", data.Username.String()))
			return
		}
	case !data.WithSelector.IsNull():
		opts := hcloud.StorageBoxSubaccountListOpts{}
		opts.LabelSelector = data.WithSelector.ValueString()

		all, err := d.client.StorageBox.AllSubaccountsWithOpts(ctx, storageBox, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		var newDiag diag.Diagnostic
		result, newDiag = datasourceutil.GetOneResultForLabelSelector("storage box subaccount", all, opts.LabelSelector)
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
