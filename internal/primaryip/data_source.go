package primaryip

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

const DataSourceType = "hcloud_primary_ip"

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Primary IP.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Primary IP.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the Primary IP (`ipv4` or `ipv6`).",
			Computed:            true,
		},
		"location": schema.StringAttribute{
			MarkdownDescription: "Name of the Location of the Primary IP.",
			Computed:            true,
		},
		"datacenter": schema.StringAttribute{
			MarkdownDescription: "Name of the Datacenter of the Primary IP.",
			Computed:            true,
			DeprecationMessage:  "The datacenter attribute is deprecated and will be removed after 1 July 2026. Please use the location attribute instead. See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters.",
		},

		"assignee_id": schema.Int64Attribute{
			MarkdownDescription: "ID of the resource the Primary IP is assigned to.",
			Computed:            true,
		},
		"assignee_type": schema.StringAttribute{
			MarkdownDescription: "Type of the resource the Primary IP is assigned to.",
			Computed:            true,
		},

		"auto_delete": schema.BoolAttribute{
			MarkdownDescription: "Whether auto delete is enabled.",
			Computed:            true,
		},
		"labels": datasourceutil.LabelsSchema(),
		"delete_protection": schema.BoolAttribute{
			MarkdownDescription: " Whether delete protection is enabled.",
			Computed:            true,
		},
		"ip_address": schema.StringAttribute{
			MarkdownDescription: "IP address of the Primary IP.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"ip_network": schema.StringAttribute{
			MarkdownDescription: "IP network of the Primary IP for IPv6 addresses. Only set if `type` is `ipv6`.",
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
Provides details about a Hetzner Cloud Primary IP.

See the [Primary IPs API documentation](https://docs.hetzner.cloud/reference/cloud#tag/primary-ips) for more details.
`

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

var _ util.ModelFromAPI[*hcloud.PrimaryIP] = &dataSourceModel{}

func (m *dataSourceModel) FromAPI(ctx context.Context, in *hcloud.PrimaryIP) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(m.model.FromAPI(ctx, in)...)

	return diags
}

func (d *DataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
			path.MatchRoot("ip_address"),
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

	var result *hcloud.PrimaryIP
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.PrimaryIP.GetByID(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("primary ip", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.PrimaryIP.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("primary ip", "name", data.Name.String()))
			return
		}
	case !data.IPAddress.IsNull():
		result, _, err = d.client.PrimaryIP.GetByIP(ctx, data.IPAddress.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("primary ip", "ip_address", data.IPAddress.String()))
			return
		}
	case !data.WithSelector.IsNull():
		opts := hcloud.PrimaryIPListOpts{}
		opts.LabelSelector = data.WithSelector.ValueString()

		all, err := d.client.PrimaryIP.AllWithOpts(ctx, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		var newDiag diag.Diagnostic
		result, newDiag = datasourceutil.GetOneResultForLabelSelector("primary ip", all, opts.LabelSelector)
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
