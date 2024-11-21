package sshkey

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

// DataSourceListType is the type name of the Hetzner Cloud SSH Keys data source.
const DataSourceListType = "hcloud_ssh_keys"

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
Provides a list of Hetzner Cloud SSH Keys.

This resource is useful if you want to use a non-terraform managed SSH Key.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional: true,
		},
		"ssh_keys": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSourceSchema(true),
			},
			Computed: true,
		},
		"with_selector": schema.StringAttribute{
			MarkdownDescription: "Filter results using a [Label Selector](https://docs.hetzner.cloud/#label-selector)",
			Optional:            true,
		},
	}
}

type resourceDataList struct {
	ID      types.String `tfsdk:"id"`
	SSHKeys types.List   `tfsdk:"ssh_keys"`

	WithSelector types.String `tfsdk:"with_selector"`
}

func populateResourceDataList(ctx context.Context, data *resourceDataList, in []*hcloud.SSHKey) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	// This type is required because the SDK version had an `int` ID field inside the data source list,
	// but a `string` ID field for the resource & single data source.
	type resourceDataList struct {
		ID          types.Int64  `tfsdk:"id"`
		Name        types.String `tfsdk:"name"`
		Fingerprint types.String `tfsdk:"fingerprint"`
		PublicKey   types.String `tfsdk:"public_key"`
		Labels      types.Map    `tfsdk:"labels"`
	}
	var populateResourceDataList = func(ctx context.Context, data *resourceDataList, in *hcloud.SSHKey) diag.Diagnostics {
		var diags diag.Diagnostics
		var newDiags diag.Diagnostics

		data.ID = types.Int64Value(int64(in.ID))
		data.Name = types.StringValue(in.Name)
		data.Fingerprint = types.StringValue(in.Fingerprint)
		data.PublicKey = types.StringValue(in.PublicKey)

		data.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, in.Labels)
		diags.Append(newDiags...)

		return diags
	}

	sshKeyIDs := make([]string, len(in))
	sshKeys := make([]resourceDataList, len(in))

	for i, item := range in {
		sshKeyIDs[i] = strconv.Itoa(item.ID)

		var sshKey resourceDataList
		diags.Append(populateResourceDataList(ctx, &sshKey, item)...)
		sshKeys[i] = sshKey
	}

	data.ID = types.StringValue(strings.Join(sshKeyIDs, "-"))
	data.SSHKeys, newDiags = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":          types.Int64Type,
		"name":        types.StringType,
		"fingerprint": types.StringType,
		"public_key":  types.StringType,
		"labels":      types.MapType{ElemType: types.StringType},
	}}, sshKeys)
	diags.Append(newDiags...)

	return diags
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

	var result []*hcloud.SSHKey
	var err error

	opts := hcloud.SSHKeyListOpts{}
	if !data.WithSelector.IsNull() {
		opts.LabelSelector = data.WithSelector.ValueString()
	}

	result, err = d.client.SSHKey.AllWithOpts(ctx, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(populateResourceDataList(ctx, &data, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
