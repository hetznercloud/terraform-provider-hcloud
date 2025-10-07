package zonerrset

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
const DataSourceType = "hcloud_zone_rrset"

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"zone": schema.StringAttribute{
			MarkdownDescription: "ID or Name of the parent Zone.",
			Required:            !readOnly,
			Optional:            readOnly,
			Computed:            readOnly,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the Zone RRSet.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Zone RRSet.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the Zone RRSet.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"ttl": schema.Int32Attribute{
			MarkdownDescription: "Time To Live (TTL) of the Zone RRSet.",
			Computed:            true,
		},
		"labels": schema.MapAttribute{
			MarkdownDescription: "User-defined [labels](https://docs.hetzner.cloud/reference/cloud#labels) (key-value pairs) for the resource.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"change_protection": schema.BoolAttribute{
			MarkdownDescription: "Whether change protection is enabled.",
			Computed:            true,
		},
		"records": schema.ListNestedAttribute{
			MarkdownDescription: "Records of the Zone RRSet.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"value": schema.StringAttribute{
						MarkdownDescription: "Value of the record.",
						Computed:            true,
					},
					"comment": schema.StringAttribute{
						MarkdownDescription: "Comment of the record.",
						Computed:            true,
					},
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
Provides details about a Hetzner Cloud Zone Resource Record Set (RRSet).

See the [Zone RRSets API documentation](https://docs.hetzner.cloud/reference/cloud#zone-rrsets) for more details.
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

func populateDataSourceModel(ctx context.Context, data *dataSourceModel, in *hcloud.ZoneRRSet) diag.Diagnostics {
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
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("name"),
			path.MatchRoot("type"),
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

	zone := &hcloud.Zone{Name: data.Zone.ValueString()}

	var result *hcloud.ZoneRRSet
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.Zone.GetRRSetByID(ctx, zone, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("zone", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull() && !data.Type.IsNull():
		result, _, err = d.client.Zone.GetRRSetByNameAndType(ctx, zone, data.Name.ValueString(), hcloud.ZoneRRSetType(data.Type.ValueString()))
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("zone", "name", data.Name.String()))
			return
		}
	case !data.WithSelector.IsNull():
		opts := hcloud.ZoneRRSetListOpts{}
		opts.LabelSelector = data.WithSelector.ValueString()

		all, err := d.client.Zone.AllRRSetsWithOpts(ctx, zone, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		var newDiag diag.Diagnostic
		result, newDiag = datasourceutil.GetOneResultForLabelSelector("zone rrset", all, opts.LabelSelector)
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
