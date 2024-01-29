package location

import (
	"context"
	"crypto/sha1"
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud location datasource.
	DataSourceType = "hcloud_location"

	// DataSourceListType is the type name of the Hetzner Cloud location datasource.
	DataSourceListType = "hcloud_locations"
)

type resourceData struct {
	ID          types.Int64   `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	Country     types.String  `tfsdk:"country"`
	City        types.String  `tfsdk:"city"`
	Latitude    types.Float64 `tfsdk:"latitude"`
	Longitude   types.Float64 `tfsdk:"longitude"`
	NetworkZone types.String  `tfsdk:"network_zone"`
}

var resourceDataAttrTypes = map[string]attr.Type{
	"id":           types.Int64Type,
	"name":         types.StringType,
	"description":  types.StringType,
	"country":      types.StringType,
	"city":         types.StringType,
	"latitude":     types.Float64Type,
	"longitude":    types.Float64Type,
	"network_zone": types.StringType,
}

func newResourceData(_ context.Context, in *hcloud.Location) (resourceData, diag.Diagnostics) { // nolint:unparam // to keep the pattern consistent between all data sources
	var data resourceData
	var diags diag.Diagnostics

	data.ID = types.Int64Value(int64(in.ID))
	data.Name = types.StringValue(in.Name)
	data.Description = types.StringValue(in.Description)
	data.Country = types.StringValue(in.Country)
	data.City = types.StringValue(in.City)
	data.Latitude = types.Float64Value(in.Latitude)
	data.Longitude = types.Float64Value(in.Longitude)
	data.NetworkZone = types.StringValue(string(in.NetworkZone))

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
		"country": schema.StringAttribute{
			Computed: true,
		},
		"city": schema.StringAttribute{
			Computed: true,
		},
		"latitude": schema.Float64Attribute{
			Computed: true,
		},
		"longitude": schema.Float64Attribute{
			Computed: true,
		},
		"network_zone": schema.StringAttribute{
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

	var result *hcloud.Location
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.Location.GetByID(ctx, int(data.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.AddError(
				"Resource not found",
				fmt.Sprintf("No location found with id %s.", data.ID.String()),
			)
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.Location.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		}
		if result == nil {
			resp.Diagnostics.AddError(
				"Resource not found",
				fmt.Sprintf("No location found with name %s.", data.Name.String()),
			)
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
		"location_ids": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use locations list instead",
			ElementType:        types.StringType,
		},
		"names": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use locations list instead",
			ElementType:        types.StringType,
		},
		"descriptions": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use locations list instead",
			ElementType:        types.StringType,
		},
		"locations": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSchema(),
			},
			Computed: true,
		},
	}

	resp.Schema.MarkdownDescription = dataSourceListMarkdownDescription
}

type resourceDataList struct {
	ID           types.String `tfsdk:"id"`
	LocationIDs  types.List   `tfsdk:"location_ids"`
	Names        types.List   `tfsdk:"names"`
	Descriptions types.List   `tfsdk:"descriptions"`
	Locations    types.List   `tfsdk:"locations"`
}

func newResourceDataList(ctx context.Context, in []*hcloud.Location) (resourceDataList, diag.Diagnostics) {
	var data resourceDataList
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	locationIDs := make([]string, len(in))
	names := make([]string, len(in))
	descriptions := make([]string, len(in))
	locations := make([]resourceData, len(in))

	for i, item := range in {
		locationIDs[i] = strconv.Itoa(item.ID)
		names[i] = item.Name
		descriptions[i] = item.Description

		location, newDiags := newResourceData(ctx, item)
		diags.Append(newDiags...)
		locations[i] = location
	}

	data.ID = types.StringValue(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(locationIDs, "")))))

	data.LocationIDs, newDiags = types.ListValueFrom(ctx, types.StringType, locationIDs)
	diags.Append(newDiags...)
	data.Names, newDiags = types.ListValueFrom(ctx, types.StringType, names)
	diags.Append(newDiags...)
	data.Descriptions, newDiags = types.ListValueFrom(ctx, types.StringType, descriptions)
	diags.Append(newDiags...)

	data.Locations, newDiags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: resourceDataAttrTypes}, locations)
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

	var result []*hcloud.Location
	var err error

	result, err = d.client.Location.All(ctx)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	data, diags := newResourceDataList(ctx, result)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
