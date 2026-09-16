package networkmember

import (
	"github.com/hashicorp/terraform-plugin-framework-nettypes/cidrtypes"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getCommonDataSourceSchema(readOnly bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "ID of the Resource attached to the Network.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "Type of the Resource attached to the Network.",
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"subnet": schema.StringAttribute{
			MarkdownDescription: "IP range of the subnet the Resource is attached to.",
			CustomType:          cidrtypes.IPPrefixType{},
			Optional:            !readOnly,
			Computed:            readOnly,
		},
		"ip": schema.StringAttribute{
			MarkdownDescription: "IP address of the Resource within the Network.",
			CustomType:          iptypes.IPAddressType{},
			Computed:            true,
		},
		"alias_ips": schema.SetAttribute{
			MarkdownDescription: "Additional IP addresses of the Resource within the Network.",
			ElementType:         iptypes.IPAddressType{},
			Computed:            true,
		},
		"status": schema.StringAttribute{
			MarkdownDescription: "Status of the Resource within the Network.",
			Computed:            true,
		},
	}
}
