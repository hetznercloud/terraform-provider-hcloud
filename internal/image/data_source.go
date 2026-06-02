package image

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Image.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the Image, for example `system`, `backup` or `snapshot`.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Image, only present when the type is `system`.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Image.",
			Computed:            true,
		},
		"labels": datasourceutil.LabelsSchema(),
		"created": schema.StringAttribute{
			MarkdownDescription: "Point in time when the Image was created (in RFC3339 format).",
			Computed:            true,
		},
		"os_flavor": schema.StringAttribute{
			MarkdownDescription: "Flavor of the operating system contained in the Image.",
			Computed:            true,
		},
		"os_version": schema.StringAttribute{
			MarkdownDescription: "Version of the operating system contained in the Image.",
			Computed:            true,
		},
		"architecture": schema.StringAttribute{
			MarkdownDescription: "CPU architecture compatible with the Image.",
			Computed:            true,
		},
		"rapid_deploy": schema.BoolAttribute{
			MarkdownDescription: "Whether the Image is optimized for a rapid deployment.",
			Computed:            true,
		},
		"deprecated": schema.StringAttribute{
			MarkdownDescription: "Point in time when the Image was marked as deprecated (in RFC3339 format).",
			Computed:            true,
		},
	}
}

const DataSourceType = "hcloud_image"

var _ datasource.DataSource = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigValidators = (*DataSource)(nil)

type DataSource struct {
	client *hcloud.Client
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceType
}

func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides details about a Hetzner Cloud Image.

It is recommended to always provide the image architecture (using ''with_architecture'').

See the [Image API documentation](https://docs.hetzner.cloud/reference/cloud#images) for more details.
`

	resp.Schema.Attributes = getCommonDataSourceSchema(false)
	maps.Copy(resp.Schema.Attributes, map[string]schema.Attribute{
		"selector": schema.StringAttribute{
			MarkdownDescription: "Filter results using a [Label Selector](https://docs.hetzner.cloud/reference/cloud#label-selector).",
			Optional:            true,
			DeprecationMessage:  "Please use the with_selector property instead.",
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
			MarkdownDescription: "Sort results by created date, and return the most recent result.",
			Optional:            true,
		},
		"include_deprecated": schema.BoolAttribute{
			MarkdownDescription: "Include deprecated images.",
			Optional:            true,
		},
	})
}

type dataSourceModel struct {
	model

	Selector          types.String `tfsdk:"selector"`
	WithSelector      types.String `tfsdk:"with_selector"`
	WithStatus        types.Set    `tfsdk:"with_status"`
	WithArchitecture  types.String `tfsdk:"with_architecture"`
	MostRecent        types.Bool   `tfsdk:"most_recent"`
	IncludeDeprecated types.Bool   `tfsdk:"include_deprecated"`
}

var _ util.ModelFromAPI[*hcloud.Image] = &dataSourceModel{}

func (m *dataSourceModel) FromAPI(ctx context.Context, in *hcloud.Image) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(m.model.FromAPI(ctx, in)...)

	return diags
}

func (d *DataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
			path.MatchRoot("selector"),
			path.MatchRoot("with_selector"),
		),
	}
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *hcloud.Image
	var err error
	var newDiag diag.Diagnostic

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.Image.GetByID(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("image", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		opts := hcloud.ImageListOpts{}
		opts.Name = data.Name.ValueString()

		resp.Diagnostics.Append(prepareImageListOpts(ctx, &opts, data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		all, _, err := d.client.Image.List(ctx, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		if !data.MostRecent.IsNull() && data.MostRecent.ValueBool() {
			// Sorting happens server side.
			result, newDiag = hcloudutil.GetFirst(all,
				hcloudutil.WithResourceName("image"),
				hcloudutil.WithUsing("name", opts.Name),
				hcloudutil.WithListOpts(opts),
			)
		} else {
			result, newDiag = hcloudutil.GetOne(all,
				hcloudutil.WithResourceName("image"),
				hcloudutil.WithUsing("name", opts.Name),
				hcloudutil.WithListOpts(opts),
			)
		}
		if newDiag != nil {
			resp.Diagnostics.Append(newDiag)
			return
		}
	case !data.WithSelector.IsNull() || !data.Selector.IsNull():
		opts := hcloud.ImageListOpts{}
		if !data.WithSelector.IsNull() {
			opts.LabelSelector = data.WithSelector.ValueString()
		} else if !data.Selector.IsNull() {
			opts.LabelSelector = data.Selector.ValueString()
		}

		resp.Diagnostics.Append(prepareImageListOpts(ctx, &opts, data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		all, _, err := d.client.Image.List(ctx, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		if !data.MostRecent.IsNull() && data.MostRecent.ValueBool() {
			// Sorting happens server side.
			result, newDiag = hcloudutil.GetFirst(all,
				hcloudutil.WithResourceName("image"),
				hcloudutil.WithUsing("label selector", opts.LabelSelector),
				hcloudutil.WithListOpts(opts),
			)
		} else {
			result, newDiag = hcloudutil.GetOne(all,
				hcloudutil.WithResourceName("image"),
				hcloudutil.WithUsing("label selector", opts.LabelSelector),
				hcloudutil.WithListOpts(opts),
			)
		}
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

func prepareImageListOpts(ctx context.Context, opts *hcloud.ImageListOpts, data dataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if !data.WithStatus.IsNull() {
		diags.Append(data.WithStatus.ElementsAs(ctx, &opts.Status, false)...)
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

	return diags
}
