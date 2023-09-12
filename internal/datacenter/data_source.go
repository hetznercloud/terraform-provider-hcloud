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
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Datacenter datasource.
	DataSourceType = "hcloud_datacenter"

	// DataSourceListType is the type name of the Hetzner Cloud Datacenters datasource.
	DataSourceListType = "hcloud_datacenters"
)

type datacenterResourceData struct {
	ID                     types.Int64  `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	Location               types.Map    `tfsdk:"location"`
	SupportedServerTypeIds types.List   `tfsdk:"supported_server_type_ids"`
	AvailableServerTypeIds types.List   `tfsdk:"available_server_type_ids"`
}

func (o *datacenterResourceData) fromResponse(ctx context.Context, in *hcloud.Datacenter) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	o.ID = types.Int64Value(int64(in.ID))
	o.Name = types.StringValue(in.Name)
	o.Description = types.StringValue(in.Description)

	o.Location, newDiags = types.MapValue(types.StringType, map[string]attr.Value{
		"id":          types.StringValue(strconv.Itoa(in.Location.ID)),
		"name":        types.StringValue(in.Location.Name),
		"description": types.StringValue(in.Location.Description),
		"country":     types.StringValue(in.Location.Country),
		"city":        types.StringValue(in.Location.City),
		"latitude":    types.StringValue(fmt.Sprintf("%f", in.Location.Latitude)),
		"longitude":   types.StringValue(fmt.Sprintf("%f", in.Location.Longitude)),
	})
	diags.Append(newDiags...)

	supportedServerTypeIds := make([]int64, len(in.ServerTypes.Supported))
	for i, v := range in.ServerTypes.Supported {
		supportedServerTypeIds[i] = int64(v.ID)
	}
	availableServerTypeIds := make([]int64, len(in.ServerTypes.Available))
	for i, v := range in.ServerTypes.Available {
		availableServerTypeIds[i] = int64(v.ID)
	}
	sort.Slice(supportedServerTypeIds, func(i, j int) bool { return supportedServerTypeIds[i] < supportedServerTypeIds[j] })
	sort.Slice(availableServerTypeIds, func(i, j int) bool { return availableServerTypeIds[i] < availableServerTypeIds[j] })

	o.SupportedServerTypeIds, newDiags = types.ListValueFrom(ctx, types.Int64Type, supportedServerTypeIds)
	diags.Append(newDiags...)
	o.AvailableServerTypeIds, newDiags = types.ListValueFrom(ctx, types.Int64Type, availableServerTypeIds)
	diags.Append(newDiags...)

	return diags
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
		// TODO: Refactor to SingleNestedAttribute in v2
		"location": schema.MapAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
		"supported_server_type_ids": schema.ListAttribute{
			Computed:    true,
			ElementType: types.Int64Type,
		},
		"available_server_type_ids": schema.ListAttribute{
			Computed:    true,
			ElementType: types.Int64Type,
		},
	}
}

// Single
var _ datasource.DataSource = (*datacenterDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*datacenterDataSource)(nil)
var _ datasource.DataSourceWithConfigValidators = (*datacenterDataSource)(nil)

type datacenterDataSource struct {
	client *hcloud.Client
}

func NewDataSource() datasource.DataSource {
	return &datacenterDataSource{}
}

// Metadata should return the full name of the data source.
func (d *datacenterDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *datacenterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// TODO: refactor to a reusable function
	client, ok := req.ProviderData.(*hcloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hcloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Schema should return the schema for this data source.
func (d *datacenterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = getCommonDataSchema()

	resp.Schema.MarkdownDescription = `
Provides details about a specific Hetzner Cloud Datacenter.
Use this resource to get detailed information about specific datacenter.

## Example Usage
` + "```" + `hcl
data "hcloud_datacenter" "ds_1" {
  name = "fsn1-dc8"
}
data "hcloud_datacenter" "ds_2" {
  id = 4
}
` + "```"
}

// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
func (d *datacenterDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
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
func (d *datacenterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datacenterResourceData

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
			resp.Diagnostics.Append(hcclient.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.AddError(
				"Resource not found",
				fmt.Sprintf("No datacenter found with id %s.", data.ID.String()),
			)
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.Datacenter.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcclient.APIErrorDiagnostics(err)...)
		}
		if result == nil {
			resp.Diagnostics.AddError(
				"Resource not found",
				fmt.Sprintf("No datacenter found with name %s.", data.Name.String()),
			)
			return
		}
	default:
		// Should not happen, see [datacenterDataSource.ConfigValidators]
		resp.Diagnostics.AddError("Unexpected internal error", "")
		return
	}

	data.fromResponse(ctx, result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// List
var _ datasource.DataSource = (*datacentersDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*datacentersDataSource)(nil)
var _ datasource.DataSourceWithConfigValidators = (*datacentersDataSource)(nil)

type datacentersDataSource struct {
	client *hcloud.Client
}

func NewListDataSource() datasource.DataSource {
	return &datacentersDataSource{}
}

// Metadata should return the full name of the data source.
func (d *datacentersDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = DataSourceListType
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *datacentersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// TODO: refactor to a reusable function
	client, ok := req.ProviderData.(*hcloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hcloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Schema should return the schema for this data source.
func (d *datacentersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Attributes = map[string]schema.Attribute{
		"datacenter_ids": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use datacenters list instead",
			ElementType:        types.StringType,
		},
		"name": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use datacenters list instead",
			ElementType:        types.StringType,
		},
		"descriptions": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use datacenters list instead",
			ElementType:        types.StringType,
		},
		"datacenters": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSchema(),
			},
			Computed: true,
		},
	}

	resp.Schema.MarkdownDescription = `
Provides details about a specific Hetzner Cloud Datacenter.
Use this resource to get detailed information about specific datacenter.

## Example Usage
` + "```" + `hcl
data "hcloud_datacenter" "ds_1" {
  name = "fsn1-dc8"
}
data "hcloud_datacenter" "ds_2" {
  id = 4
}
` + "```"
}

// ConfigValidators returns a list of ConfigValidators. Each ConfigValidator's Validate method will be called when validating the data source.
func (d *datacentersDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

type datacentersResourceData struct {
	DatacenterIDs []types.String           `tfsdk:"datacenter_ids"`
	Names         []types.String           `tfsdk:"names"`
	Descriptions  []types.String           `tfsdk:"descriptions"`
	Datacenters   []datacenterResourceData `tfsdk:"datacenters"`
}

// Read is called when the provider must read data source values in
// order to update state. Config values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (d *datacentersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datacentersResourceData

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result []*hcloud.Datacenter
	var err error

	// 	dcs, err := client.Datacenter.All(ctx)
	// 	if err != nil {
	// 		return hcclient.ErrorToDiag(err)
	// 	}

	// func dataSourceHcloudDatacenterListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// 	client := m.(*hcloud.Client)
	// 	dcs, err := client.Datacenter.All(ctx)
	// 	if err != nil {
	// 		return hcclient.ErrorToDiag(err)
	// 	}

	// 	names := make([]string, len(dcs))
	// 	descriptions := make([]string, len(dcs))
	// 	ids := make([]string, len(dcs))
	// 	tfDatacenters := make([]map[string]interface{}, len(dcs))
	// 	for i, datacenter := range dcs {
	// 		ids[i] = strconv.Itoa(datacenter.ID)
	// 		descriptions[i] = datacenter.Description
	// 		names[i] = datacenter.Name

	// 		tfDatacenters[i] = getDatacenterAttributes(datacenter)
	// 	}
	// 	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))
	// 	d.Set("datacenter_ids", ids)
	// 	d.Set("names", names)
	// 	d.Set("descriptions", descriptions)
	// 	d.Set("datacenters", tfDatacenters)

	// 	return nil
	// }

	switch {
	default:
		// Should not happen, see [datacenterDataSource.ConfigValidators]
		resp.Diagnostics.AddError("Unexpected internal error", "")
		return
	}

	data.fromResponse(ctx, result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
