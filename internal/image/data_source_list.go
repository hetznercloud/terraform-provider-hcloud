package image

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

// DataSourceListType is the type name of the Hetzner Cloud Images datasource.
const DataSourceListType = "hcloud_images"

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
Provides a list of Hetzner Storage Images.

It is recommended to always provide the image architecture (using ''with_architecture'').

See the [Image API documentation](https://docs.hetzner.cloud/reference/cloud#images) for more details.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"images": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSourceSchema(true),
			},
			Computed: true,
		},
		"with_selector": schema.StringAttribute{
			MarkdownDescription: "Filter results using a [Label Selector](https://docs.hetzner.cloud/reference/hetzner#label-selector).",
			Optional:            true,
		},
		"with_status": schema.SetAttribute{
			MarkdownDescription: "Filter results by statuses, for example `creating` or `available`.",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"with_architecture": schema.StringAttribute{
			MarkdownDescription: "Filter results by architecture, for example `x86` (default) or `arm`.",
			Optional:            true,
		},
		"most_recent": schema.BoolAttribute{
			MarkdownDescription: "Sort results by created date.",
			Optional:            true,
		},
		"include_deprecated": schema.BoolAttribute{
			MarkdownDescription: "Include deprecated images.",
			Optional:            true,
		},
	}
}

type dataSourceListModel struct {
	ID     types.String `tfsdk:"id"`
	Images types.List   `tfsdk:"images"`

	WithSelector      types.String `tfsdk:"with_selector"`
	WithStatus        types.Set    `tfsdk:"with_status"`
	WithArchitecture  types.String `tfsdk:"with_architecture"`
	MostRecent        types.Bool   `tfsdk:"most_recent"`
	IncludeDeprecated types.Bool   `tfsdk:"include_deprecated"`
}

var _ util.ModelFromAPI[[]*hcloud.Image] = &dataSourceListModel{}

func (m *dataSourceListModel) FromAPI(ctx context.Context, in []*hcloud.Image) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	tfIDs := make([]string, 0, len(in))
	tfItems := make([]attr.Value, 0, len(in))
	for _, item := range in {
		var value model
		diags.Append(value.FromAPI(ctx, item)...)

		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
		tfIDs = append(tfIDs, util.FormatID(item.ID))
	}

	m.ID = types.StringValue(datasourceutil.ListID(tfIDs))
	m.Images, newDiags = types.ListValue((&model{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return diags
}

func (d *DataSourceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceListModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result []*hcloud.Image
	var err error

	opts := hcloud.ImageListOpts{}
	if !data.WithSelector.IsNull() {
		opts.LabelSelector = data.WithSelector.ValueString()
	}

	if !data.WithStatus.IsNull() {
		resp.Diagnostics.Append(data.WithStatus.ElementsAs(ctx, &opts.Status, false)...)
		return
	}

	if !data.WithArchitecture.IsNull() {
		opts.Architecture = []hcloud.Architecture{
			hcloud.Architecture(data.WithArchitecture.ValueString()),
		}
	} else {
		opts.Architecture = []hcloud.Architecture{
			hcloud.ArchitectureX86,
		}
	}

	if !data.IncludeDeprecated.IsNull() {
		opts.IncludeDeprecated = data.IncludeDeprecated.ValueBool()
	}

	if !data.MostRecent.IsNull() && data.MostRecent.ValueBool() {
		opts.Sort = []string{"created:desc"}
	}

	result, err = d.client.Image.AllWithOpts(ctx, opts)
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
