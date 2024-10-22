package loadbalancertype

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud load balancer type datasource.
	DataSourceType = "hcloud_load_balancer_type"

	// DataSourceListType is the type name of the Hetzner Cloud load balancer type list datasource.
	DataSourceListType = "hcloud_load_balancer_types"
)

type resourceData struct {
	ID                      types.Int64  `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	MaxAssignedCertificates types.Int64  `tfsdk:"max_assigned_certificates"`
	MaxConnections          types.Int64  `tfsdk:"max_connections"`
	MaxServices             types.Int64  `tfsdk:"max_services"`
	MaxTargets              types.Int64  `tfsdk:"max_targets"`
}

var resourceDataAttrTypes = map[string]attr.Type{
	"id":                        types.Int64Type,
	"name":                      types.StringType,
	"description":               types.StringType,
	"max_assigned_certificates": types.Int64Type,
	"max_connections":           types.Int64Type,
	"max_services":              types.Int64Type,
	"max_targets":               types.Int64Type,
}

func newResourceData(_ context.Context, in *hcloud.LoadBalancerType) (resourceData, diag.Diagnostics) { //nolint:unparam
	var data resourceData
	var diags diag.Diagnostics

	data.ID = types.Int64Value(int64(in.ID))
	data.Name = types.StringValue(in.Name)
	data.Description = types.StringValue(in.Description)

	data.MaxAssignedCertificates = types.Int64Value(int64(in.MaxAssignedCertificates))
	data.MaxConnections = types.Int64Value(int64(in.MaxConnections))
	data.MaxServices = types.Int64Value(int64(in.MaxServices))
	data.MaxTargets = types.Int64Value(int64(in.MaxTargets))

	return data, diags
}

func getCommonDataSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"name": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"description": schema.StringAttribute{
			Computed: true,
		},
		"max_assigned_certificates": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_connections": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_services": schema.Int64Attribute{
			Optional: true,
			Computed: true,
		},
		"max_targets": schema.Int64Attribute{
			Optional: true,
			Computed: true,
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
	resp.Schema.Attributes = getCommonDataSchema()
	resp.Schema.MarkdownDescription = dataSourceMarkdownDescription
}

// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
func (d *dataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resourceData

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result *hcloud.LoadBalancerType
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.LoadBalancerType.GetByID(ctx, int(data.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load balancer type", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.LoadBalancerType.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("load balancer type", "name", data.Name.String()))
			return
		}
	default:
		// Should not happen, see [dataSource.ConfigValidators]
		resp.Diagnostics.AddError("Unexpected internal error", "")
		return
	}

	data, diags := newResourceData(ctx, result)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// List
var _ datasource.DataSource = (*dataSourceList)(nil)
var _ datasource.DataSourceWithConfigure = (*dataSourceList)(nil)

type dataSourceList struct {
	client *hcloud.Client
}

func NewDataSourceList() datasource.DataSource {
	return &dataSourceList{}
}

// Metadata should return the full name of the data source.
func (d *dataSourceList) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceListType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSourceList) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	var newDiags diag.Diagnostics

	d.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//go:embed data_source_list.md
var dataSourceListMarkdownDescription string

// Schema should return the schema for this data source.
func (d *dataSourceList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional: true,
		},
		"load_balancer_types": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSchema(),
			},
			Computed: true,
		},
	}

	resp.Schema.MarkdownDescription = dataSourceListMarkdownDescription
}

type resourceDataList struct {
	ID                types.String `tfsdk:"id"`
	LoadBalancerTypes types.List   `tfsdk:"load_balancer_types"`
}

func newResourceDataList(ctx context.Context, in []*hcloud.LoadBalancerType) (resourceDataList, diag.Diagnostics) {
	var data resourceDataList
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	ids := make([]int64, len(in))
	tfItems := make([]resourceData, len(in))
	for i, item := range in {
		ids[i] = int64(item.ID)

		tfItem, newDiags := newResourceData(ctx, item)
		diags.Append(newDiags...)
		tfItems[i] = tfItem
	}

	data.ID = types.StringValue(datasourceutil.ListID(ids))

	data.LoadBalancerTypes, newDiags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: resourceDataAttrTypes}, tfItems)
	diags.Append(newDiags...)

	return data, diags
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resourceDataList

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result []*hcloud.LoadBalancerType
	var err error

	result, err = d.client.LoadBalancerType.All(ctx)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	data, diags := newResourceDataList(ctx, result)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
