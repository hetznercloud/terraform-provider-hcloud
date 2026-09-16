package networkmember

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/sliceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const DataSourceListType = "hcloud_network_members"

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
	resp.Schema.MarkdownDescription = util.MarkdownDescription(`
Provides a list of Hetzner Cloud Network Members.

See the [Networks API documentation](https://docs.hetzner.cloud/reference/cloud#tag/networks) for more details.
`)

	resp.Schema.Attributes = map[string]schema.Attribute{
		"network_id": schema.Int64Attribute{
			MarkdownDescription: "ID the Network to list the members of.",
			Required:            true,
		},
		"members": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: getCommonDataSourceSchema(true),
			},
			Computed: true,
		},
		"with_type": schema.SetAttribute{
			MarkdownDescription: "Filter members by type. May be used multiple times.",
			ElementType:         types.StringType,
			Optional:            true,
			Validators:          []validator.Set{setvalidator.ValueStringsAre(stringvalidator.OneOf("server", "load_balancer"))},
		},
		"with_status": schema.SetAttribute{
			MarkdownDescription: "Filter members by status. May be used multiple times.",
			ElementType:         types.StringType,
			Optional:            true,
			Validators:          []validator.Set{setvalidator.ValueStringsAre(stringvalidator.OneOf("ok", "attaching", "detaching", "updating", "error"))},
		},
		"with_subnet": schema.SetAttribute{
			MarkdownDescription: "Filter members by subnet. May be used multiple times.",
			ElementType:         cidrtypes.IPPrefixType{},
			Optional:            true,
		},
	}
}

type dataSourceListModel struct {
	NetworkID types.Int64 `tfsdk:"network_id"`
	Members   types.List  `tfsdk:"members"`

	WithType   types.Set `tfsdk:"with_type"`
	WithStatus types.Set `tfsdk:"with_status"`
	WithSubnet types.Set `tfsdk:"with_subnet"`
}

var _ util.ModelFromAPI[[]*hcloud.NetworkMember] = &dataSourceListModel{}

func (m *dataSourceListModel) FromAPI(ctx context.Context, in []*hcloud.NetworkMember) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	tfItems := make([]attr.Value, 0, len(in))
	for _, item := range in {
		var value model
		diags.Append(value.FromAPI(ctx, item)...)

		tfItem, newDiags := value.ToTerraform(ctx)
		diags.Append(newDiags...)

		tfItems = append(tfItems, tfItem)
	}

	m.Members, newDiags = types.ListValue((&model{}).tfType(), tfItems)
	diags.Append(newDiags...)

	return diags
}

func (d *DataSourceList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceListModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result []*hcloud.NetworkMember
	var err error

	opts := hcloud.NetworkMemberListOpts{}

	if !data.WithStatus.IsNull() {
		values := make([]string, 0, len(data.WithStatus.Elements()))
		resp.Diagnostics.Append(data.WithStatus.ElementsAs(ctx, &values, false)...)

		opts.Status = sliceutil.Transform(values, func(o string) hcloud.NetworkMemberStatus { return hcloud.NetworkMemberStatus(o) })
	}

	if !data.WithType.IsNull() {
		values := make([]string, 0, len(data.WithType.Elements()))
		resp.Diagnostics.Append(data.WithType.ElementsAs(ctx, &values, false)...)

		opts.Type = sliceutil.Transform(values, func(o string) hcloud.NetworkMemberType { return hcloud.NetworkMemberType(o) })
	}

	if !data.WithSubnet.IsNull() {
		values := make([]string, 0, len(data.WithSubnet.Elements()))
		resp.Diagnostics.Append(data.WithSubnet.ElementsAs(ctx, &values, false)...)

		opts.Subnet = sliceutil.Transform(values, func(o string) *net.IPNet {
			_, value, _ := net.ParseCIDR(o)
			return value
		})
	}

	result, err = d.client.Network.AllMembersWithOpts(ctx, &hcloud.Network{ID: data.NetworkID.ValueInt64()}, opts)
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
