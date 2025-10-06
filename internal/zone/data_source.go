package zone

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceType is the type name of the Hetzner Cloud Zone data source.
const DataSourceType = "hcloud_zone"

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Zone.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Zone.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"mode": schema.StringAttribute{
			MarkdownDescription: "Mode of the Zone.",
			Computed:            true,
		},
		"ttl": schema.Int32Attribute{
			MarkdownDescription: "Default Time To Live (TTL) of the Zone.",
			Computed:            true,
		},
		"labels": schema.MapAttribute{
			MarkdownDescription: "User-defined [labels](https://docs.hetzner.cloud/reference/cloud#labels) (key-value pairs) for the resource.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"delete_protection": schema.BoolAttribute{
			MarkdownDescription: "Whether delete protection is enabled.",
			Computed:            true,
		},
		"primary_nameservers": schema.ListNestedAttribute{
			MarkdownDescription: "Primary nameservers of the Zone.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "Public IPv4 or IPv6 address of the primary nameserver.",
						Computed:            true,
					},
					"port": schema.Int32Attribute{
						MarkdownDescription: "Port of the primary nameserver.",
						Computed:            true,
					},
					"tsig_algorithm": schema.StringAttribute{
						MarkdownDescription: "Transaction signature (TSIG) algorithm used to generate the TSIG key.",
						Computed:            true,
					},
					"tsig_key": schema.StringAttribute{
						MarkdownDescription: "Transaction signature (TSIG) key",
						Computed:            true,
					},
				},
			},
		},
		"authoritative_nameservers": schema.SingleNestedAttribute{
			MarkdownDescription: "Authoritative nameservers of the Zone.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"assigned": schema.ListAttribute{
					MarkdownDescription: "Authoritative Hetzner nameservers assigned to the Zone.",
					ElementType:         types.StringType,
					Computed:            true,
				},
			},
		},
		"registrar": schema.StringAttribute{
			MarkdownDescription: "Registrar of the Zone.",
			Computed:            true,
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
	experimental.DNS.AppendDiagnostic(&resp.Diagnostics)

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
Provides details about a Hetzner Cloud Zone.

For Internationalized domain names (IDN), see the ` + "`provider::hcloud::idna`" + ` function.

See the [Zones API documentation](https://docs.hetzner.cloud/reference/cloud#zones) for more details.
`

	experimental.DNS.AppendNotice(&resp.Schema.MarkdownDescription)

	resp.Schema.Attributes = getCommonDataSourceSchema(false)
	maps.Copy(resp.Schema.Attributes, map[string]schema.Attribute{
		"with_selector": schema.StringAttribute{
			MarkdownDescription: "Filter results using a [Label Selector](https://docs.hetzner.cloud/reference/cloud#label-selector).",
			Optional:            true,
		},
	})
}

type dataSourceModel struct {
	model

	WithSelector types.String `tfsdk:"with_selector"`
}

func populateDataSourceModel(ctx context.Context, data *dataSourceModel, in *hcloud.Zone) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(data.model.FromAPI(ctx, in)...)

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

	var result *hcloud.Zone
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.Zone.GetByID(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("zone", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.Zone.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("zone", "name", data.Name.String()))
			return
		}
	case !data.WithSelector.IsNull():
		opts := hcloud.ZoneListOpts{}
		opts.LabelSelector = data.WithSelector.ValueString()

		all, err := d.client.Zone.AllWithOpts(ctx, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		var newDiag diag.Diagnostic
		result, newDiag = datasourceutil.GetOneResultForLabelSelector("zone", all, opts.LabelSelector)
		if newDiag != nil {
			resp.Diagnostics.Append(newDiag)
			return
		}
	}

	resp.Diagnostics.Append(populateDataSourceModel(ctx, &data, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
