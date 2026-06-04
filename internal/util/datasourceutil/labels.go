package datasourceutil

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// LabelsSchema returns a map attribute schema for the labels field shared by multiple data sources.
func LabelsSchema() schema.MapAttribute {
	return schema.MapAttribute{
		MarkdownDescription: "User-defined [labels](https://docs.hetzner.cloud/reference/cloud#labels) (key-value pairs) for the resource.",
		Computed:            true,
		ElementType:         types.StringType,
	}
}
