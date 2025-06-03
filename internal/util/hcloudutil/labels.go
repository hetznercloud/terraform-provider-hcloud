package hcloudutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TerraformLabelsToHCloud(ctx context.Context, inputLabels types.Map, outputLabels *map[string]string) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	*outputLabels = make(map[string]string, len(inputLabels.Elements()))

	if !inputLabels.IsUnknown() && !inputLabels.IsNull() {
		diagnostics.Append(inputLabels.ElementsAs(ctx, outputLabels, false)...)
	}

	return diagnostics
}
