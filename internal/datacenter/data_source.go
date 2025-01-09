package datacenter

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Datacenter datasource.
	DataSourceType = "hcloud_datacenter"

	// DataSourceListType is the type name of the Hetzner Cloud Datacenters datasource.
	DataSourceListType = "hcloud_datacenters"
)

type resourceData struct {
	ID                     types.Int64  `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Location               types.Map    `tfsdk:"location"`
	SupportedServerTypeIDs types.List   `tfsdk:"supported_server_type_ids"`
	AvailableServerTypeIDs types.List   `tfsdk:"available_server_type_ids"`
}

var resourceDataAttrTypes = map[string]attr.Type{
	"id":                        types.Int64Type,
	"name":                      types.StringType,
	"description":               types.StringType,
	"location":                  types.MapType{ElemType: types.StringType},
	"supported_server_type_ids": types.ListType{ElemType: types.Int64Type},
	"available_server_type_ids": types.ListType{ElemType: types.Int64Type},
}

func newResourceData(ctx context.Context, in *hcloud.Datacenter) (resourceData, diag.Diagnostics) {
	var data resourceData
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	data.ID = types.Int64Value(int64(in.ID))
	data.Name = types.StringValue(in.Name)
	data.Description = types.StringValue(in.Description)

	data.Location, newDiags = types.MapValue(types.StringType, map[string]attr.Value{
		"id":          types.StringValue(strconv.Itoa(in.Location.ID)),
		"name":        types.StringValue(in.Location.Name),
		"description": types.StringValue(in.Location.Description),
		"country":     types.StringValue(in.Location.Country),
		"city":        types.StringValue(in.Location.City),
		"latitude":    types.StringValue(fmt.Sprintf("%f", in.Location.Latitude)),
		"longitude":   types.StringValue(fmt.Sprintf("%f", in.Location.Longitude)),
	})
	diags.Append(newDiags...)

	supportedServerTypeIDs := make([]int64, len(in.ServerTypes.Supported))
	for i, v := range in.ServerTypes.Supported {
		supportedServerTypeIDs[i] = int64(v.ID)
	}
	availableServerTypeIDs := make([]int64, len(in.ServerTypes.Available))
	for i, v := range in.ServerTypes.Available {
		availableServerTypeIDs[i] = int64(v.ID)
	}
	sort.Slice(supportedServerTypeIDs, func(i, j int) bool { return supportedServerTypeIDs[i] < supportedServerTypeIDs[j] })
	sort.Slice(availableServerTypeIDs, func(i, j int) bool { return availableServerTypeIDs[i] < availableServerTypeIDs[j] })

	data.SupportedServerTypeIDs, newDiags = types.ListValueFrom(ctx, types.Int64Type, supportedServerTypeIDs)
	diags.Append(newDiags...)
	data.AvailableServerTypeIDs, newDiags = types.ListValueFrom(ctx, types.Int64Type, availableServerTypeIDs)
	diags.Append(newDiags...)

	return data, diags
}

func getCommonDataSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Datacenter.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the Datacenter.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the Datacenter.",
			Computed:            true,
		},
		// TODO: Refactor to SingleNestedAttribute in v2
		"location": schema.MapAttribute{
			MarkdownDescription: "Location of the Datacenter. See the [Hetzner Docs](https://docs.hetzner.com/cloud/general/locations/#what-locations-are-there) for more details about locations.",
			ElementType:         types.StringType,
			Computed:            true,
		},
		"supported_server_type_ids": schema.ListAttribute{
			MarkdownDescription: "List of supported Server Types in the Datacenter.",
			ElementType:         types.Int64Type,
			Computed:            true,
		},
		"available_server_type_ids": schema.ListAttribute{
			MarkdownDescription: "List of currently available Server Types in the Datacenter.",
			ElementType:         types.Int64Type,
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

// Schema should return the schema for this data source.
func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides details about a specific Hetzner Cloud Datacenter.

Use this resource to get detailed information about a specific Datacenter.
`
	resp.Schema.Attributes = getCommonDataSchema(false)
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

	var result *hcloud.Datacenter
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.Datacenter.GetByID(ctx, int(data.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("datacenter", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.Datacenter.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("datacenter", "name", data.Name.String()))
			return
		}
	default:
		// Should not happen, see [datacenterDataSource.ConfigValidators]
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

// Schema should return the schema for this data source.
func (d *dataSourceList) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Provides a list of available Hetzner Cloud Datacenters.

This resource may be useful to create highly available infrastructure, distributed across several Datacenters.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"datacenter_ids": schema.ListAttribute{
			DeprecationMessage: "Use datacenters list instead",
			ElementType:        types.StringType,
			Computed:           true,
		},
		"names": schema.ListAttribute{
			DeprecationMessage: "Use datacenters list instead",
			ElementType:        types.StringType,
			Computed:           true,
		},
		"descriptions": schema.ListAttribute{
			DeprecationMessage: "Use datacenters list instead",
			ElementType:        types.StringType,
			Computed:           true,
		},
		"datacenters": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSchema(true),
			},
			Computed: true,
		},
	}
}

type resourceDataList struct {
	ID            types.String `tfsdk:"id"`
	DatacenterIDs types.List   `tfsdk:"datacenter_ids"`
	Names         types.List   `tfsdk:"names"`
	Descriptions  types.List   `tfsdk:"descriptions"`
	Datacenters   types.List   `tfsdk:"datacenters"`
}

func newResourceDataList(ctx context.Context, in []*hcloud.Datacenter) (resourceDataList, diag.Diagnostics) {
	var data resourceDataList
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	datacenterIDs := make([]string, len(in))
	names := make([]string, len(in))
	descriptions := make([]string, len(in))
	datacenters := make([]resourceData, len(in))

	for i, item := range in {
		datacenterIDs[i] = strconv.Itoa(item.ID)
		names[i] = item.Name
		descriptions[i] = item.Description

		datacenter, newDiags := newResourceData(ctx, item)
		diags.Append(newDiags...)
		datacenters[i] = datacenter
	}

	data.ID = types.StringValue(datasourceutil.ListID(datacenterIDs))

	data.DatacenterIDs, newDiags = types.ListValueFrom(ctx, types.StringType, datacenterIDs)
	diags.Append(newDiags...)
	data.Names, newDiags = types.ListValueFrom(ctx, types.StringType, names)
	diags.Append(newDiags...)
	data.Descriptions, newDiags = types.ListValueFrom(ctx, types.StringType, descriptions)
	diags.Append(newDiags...)

	data.Datacenters, newDiags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: resourceDataAttrTypes}, datacenters)
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

	var result []*hcloud.Datacenter
	var err error

	result, err = d.client.Datacenter.All(ctx)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	data, diags := newResourceDataList(ctx, result)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
