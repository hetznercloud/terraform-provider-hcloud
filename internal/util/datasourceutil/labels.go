package datasourceutil

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetOneResultForLabelSelector verifies that only one item is returned for a label selector. If none or >1 are returned
// it returns an error [diag.Diagnostic].
func GetOneResultForLabelSelector[T any](resourceName string, items []*T, labelSelector string) (*T, diag.Diagnostic) {
	if len(items) == 0 {
		return nil, diag.NewErrorDiagnostic(fmt.Sprintf("No %s found for label selector", resourceName), fmt.Sprintf(
			"No %s found for label selector.\n\n"+
				"Label selector: %s\n",
			resourceName,
			labelSelector,
		))
	}
	if len(items) > 1 {
		return nil, diag.NewErrorDiagnostic(
			fmt.Sprintf("More than one %s found for label selector", resourceName),
			fmt.Sprintf(
				"More than one %s found for label selector.\n\n"+
					"Label selector: %s\n",
				resourceName,
				labelSelector,
			),
		)
	}

	return items[0], nil
}

// LabelsSchema returns a map attribute schema for the labels field shared by multiple data sources.
func LabelsSchema() schema.MapAttribute {
	return schema.MapAttribute{
		MarkdownDescription: "User-defined [labels](https://docs.hetzner.cloud/reference/cloud#labels) (key-value pairs) for the resource.",
		Computed:            true,
		ElementType:         types.StringType,
	}
}
