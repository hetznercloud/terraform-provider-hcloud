package sshkey

import (
	"context"
	_ "embed"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// DataSourceType is the type name of the Hetzner Cloud SSH Key data source.
const DataSourceType = "hcloud_ssh_key"

type resourceDataWithSelector struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	PublicKey   types.String `tfsdk:"public_key"`
	Labels      types.Map    `tfsdk:"labels"`

	Selector     types.String `tfsdk:"selector"`
	WithSelector types.String `tfsdk:"with_selector"`
}

func populateResourceDataWithSelector(ctx context.Context, data *resourceDataWithSelector, in *hcloud.SSHKey) diag.Diagnostics {
	var diags diag.Diagnostics

	var resourceDataWithoutSelector resourceData
	diags.Append(populateResourceData(ctx, &resourceDataWithoutSelector, in)...)

	data.ID = types.Int64Value(int64(in.ID))
	data.Name = resourceDataWithoutSelector.Name
	data.Fingerprint = resourceDataWithoutSelector.Fingerprint
	data.PublicKey = resourceDataWithoutSelector.PublicKey
	data.Labels = resourceDataWithoutSelector.Labels

	return diags
}

func getCommonDataSourceSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the SSH key.",
			Optional:            true,
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the SSH key.",
			Optional:            true,
			Computed:            true,
		},
		"fingerprint": schema.StringAttribute{
			MarkdownDescription: "Fingerprint of the SSH key.",
			Optional:            true,
			Computed:            true,
		},
		"public_key": schema.StringAttribute{
			MarkdownDescription: "Public key of the SSH key pair.",
			Optional:            true,
			Computed:            true,
		},
		"labels": schema.MapAttribute{
			MarkdownDescription: "User-defined [labels](https://docs.hetzner.cloud/#labels) (key-value pairs) for the resource.",
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
		},
	}
}

// Single
var _ datasource.DataSource = (*dataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataSource)(nil)
var _ datasource.DataSourceWithConfigValidators = (*dataSource)(nil)

type dataSource struct {
	client *hcloud.Client
}

func NewDataSource() datasource.DataSource {
	return &dataSource{}
}

// Metadata should return the full name of the data source.
func (d *dataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//go:embed data_source.md
var dataSourceMarkdownDescription string

// Schema should return the schema for this data source.
func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = getCommonDataSourceSchema()
	maps.Copy(resp.Schema.Attributes, map[string]schema.Attribute{
		"selector": schema.StringAttribute{
			Optional:           true,
			DeprecationMessage: "Please use the with_selector property instead.",
		},
		"with_selector": schema.StringAttribute{
			Optional: true,
		},
	})
	resp.Schema.MarkdownDescription = dataSourceMarkdownDescription
}

// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
func (d *dataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
			path.MatchRoot("fingerprint"),
			path.MatchRoot("selector"),
			path.MatchRoot("with_selector"),
		),
	}
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resourceDataWithSelector

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *hcloud.SSHKey
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.SSHKey.GetByID(ctx, int(data.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("ssh key", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.SSHKey.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("ssh key", "name", data.Name.String()))
			return
		}
	case !data.Fingerprint.IsNull():
		result, _, err = d.client.SSHKey.GetByFingerprint(ctx, data.Fingerprint.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("ssh key", "fingerprint", data.Fingerprint.String()))
			return
		}
	case !data.WithSelector.IsNull() || !data.Selector.IsNull():
		opts := hcloud.SSHKeyListOpts{}

		if !data.WithSelector.IsNull() {
			opts.LabelSelector = data.WithSelector.ValueString()
		} else if !data.Selector.IsNull() {
			opts.LabelSelector = data.Selector.ValueString()
		}

		allKeys, err := d.client.SSHKey.AllWithOpts(ctx, opts)
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}

		var newDiag diag.Diagnostic
		result, newDiag = datasourceutil.GetOneResultForLabelSelector("ssh key", allKeys, opts.LabelSelector)
		if newDiag != nil {
			resp.Diagnostics.Append(newDiag)
			return
		}
	}

	resp.Diagnostics.Append(populateResourceDataWithSelector(ctx, &data, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
