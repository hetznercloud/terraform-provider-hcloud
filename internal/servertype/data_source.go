package servertype

import (
	"context"
	"crypto/sha1"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	// DataSourceType is the type name of the Hetzner Cloud server type datasource.
	DataSourceType = "hcloud_server_type"

	// DataSourceListType is the type name of the Hetzner Cloud server type list datasource.
	DataSourceListType = "hcloud_server_types"
)

type resourceData struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Cores           types.Int32  `tfsdk:"cores"`
	Memory          types.Int32  `tfsdk:"memory"`
	Disk            types.Int32  `tfsdk:"disk"`
	StorageType     types.String `tfsdk:"storage_type"`
	CPUType         types.String `tfsdk:"cpu_type"`
	Architecture    types.String `tfsdk:"architecture"`
	IncludedTraffic types.Int64  `tfsdk:"included_traffic"`

	IsDeprecated         types.Bool   `tfsdk:"is_deprecated"`
	DeprecationAnnounced types.String `tfsdk:"deprecation_announced"`
	UnavailableAfter     types.String `tfsdk:"unavailable_after"`
}

var resourceDataAttrTypes = map[string]attr.Type{
	"id":               types.Int64Type,
	"name":             types.StringType,
	"description":      types.StringType,
	"cores":            types.Int32Type,
	"memory":           types.Int32Type,
	"disk":             types.Int32Type,
	"storage_type":     types.StringType,
	"cpu_type":         types.StringType,
	"architecture":     types.StringType,
	"included_traffic": types.Int64Type,

	"is_deprecated":         types.BoolType,
	"deprecation_announced": types.StringType,
	"unavailable_after":     types.StringType,
}

func newResourceData(_ context.Context, in *hcloud.ServerType) (resourceData, diag.Diagnostics) { // nolint:unparam // to keep the pattern consistent between all data sources
	var data resourceData
	var diags diag.Diagnostics

	data.ID = types.Int64Value(int64(in.ID))
	data.Name = types.StringValue(in.Name)
	data.Description = types.StringValue(in.Description)
	data.Cores = types.Int32Value(int32(in.Cores))
	data.Memory = types.Int32Value(int32(in.Memory))
	data.Disk = types.Int32Value(int32(in.Disk))
	data.StorageType = types.StringValue(string(in.StorageType))
	data.CPUType = types.StringValue(string(in.CPUType))
	data.Architecture = types.StringValue(string(in.Architecture))
	data.IncludedTraffic = types.Int64Value(in.IncludedTraffic)

	if in.IsDeprecated() {
		data.IsDeprecated = types.BoolValue(true)
		data.DeprecationAnnounced = types.StringValue(in.DeprecationAnnounced().Format(time.RFC3339))
		data.UnavailableAfter = types.StringValue(in.UnavailableAfter().Format(time.RFC3339))
	} else {
		data.IsDeprecated = types.BoolValue(false)
		data.DeprecationAnnounced = types.StringNull()
		data.UnavailableAfter = types.StringNull()
	}

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
		"cores": schema.Int32Attribute{
			Computed: true,
		},
		"memory": schema.Int32Attribute{
			Computed: true,
		},
		"disk": schema.Int32Attribute{
			Computed: true,
		},
		"storage_type": schema.StringAttribute{
			Computed: true,
		},
		"cpu_type": schema.StringAttribute{
			Computed: true,
		},
		"architecture": schema.StringAttribute{
			Computed: true,
		},
		"included_traffic": schema.Int64Attribute{
			Computed: true,
		},
		"is_deprecated": schema.BoolAttribute{
			Computed: true,
		},
		"deprecation_announced": schema.StringAttribute{
			Computed: true,
			Optional: true,
		},
		"unavailable_after": schema.StringAttribute{
			Computed: true,
			Optional: true,
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

	var result *hcloud.ServerType
	var err error

	switch {
	case !data.ID.IsNull():
		result, _, err = d.client.ServerType.GetByID(ctx, int(data.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("server type", "id", data.ID.String()))
			return
		}
	case !data.Name.IsNull():
		result, _, err = d.client.ServerType.GetByName(ctx, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
			return
		}
		if result == nil {
			resp.Diagnostics.Append(hcloudutil.NotFoundDiagnostic("server type", "name", data.Name.String()))
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
		"server_type_ids": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use server_types list instead",
			ElementType:        types.StringType,
		},
		"names": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use server_types list instead",
			ElementType:        types.StringType,
		},
		"descriptions": schema.ListAttribute{
			Optional:           true,
			DeprecationMessage: "Use server_types list instead",
			ElementType:        types.StringType,
		},
		"server_types": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSchema(),
			},
			Computed: true,
		},
	}

	resp.Schema.MarkdownDescription = dataSourceListMarkdownDescription
}

type resourceDataList struct {
	ID            types.String `tfsdk:"id"`
	ServerTypeIDs types.List   `tfsdk:"server_type_ids"`
	Names         types.List   `tfsdk:"names"`
	Descriptions  types.List   `tfsdk:"descriptions"`
	ServerTypes   types.List   `tfsdk:"server_types"`
}

func newResourceDataList(ctx context.Context, in []*hcloud.ServerType) (resourceDataList, diag.Diagnostics) {
	var data resourceDataList
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	serverTypeIDs := make([]string, len(in))
	names := make([]string, len(in))
	descriptions := make([]string, len(in))
	serverTypes := make([]resourceData, len(in))

	for i, item := range in {
		serverTypeIDs[i] = strconv.Itoa(item.ID)
		names[i] = item.Name
		descriptions[i] = item.Description

		location, newDiags := newResourceData(ctx, item)
		diags.Append(newDiags...)
		serverTypes[i] = location
	}

	data.ID = types.StringValue(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(serverTypeIDs, "")))))

	data.ServerTypeIDs, newDiags = types.ListValueFrom(ctx, types.StringType, serverTypeIDs)
	diags.Append(newDiags...)
	data.Names, newDiags = types.ListValueFrom(ctx, types.StringType, names)
	diags.Append(newDiags...)
	data.Descriptions, newDiags = types.ListValueFrom(ctx, types.StringType, descriptions)
	diags.Append(newDiags...)

	data.ServerTypes, newDiags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: resourceDataAttrTypes}, serverTypes)
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

	var result []*hcloud.ServerType
	var err error

	result, err = d.client.ServerType.All(ctx)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	data, diags := newResourceDataList(ctx, result)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
