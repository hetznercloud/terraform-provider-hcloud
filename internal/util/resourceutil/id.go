package resourceutil

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

func ParseID(value types.String) (int64, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	id, err := util.ParseID(value.ValueString())
	if err != nil {
		diagnostics.AddAttributeError(
			path.Root("id"),
			"Could not parse resource ID",
			err.Error(),
		)
		return 0, diagnostics
	}

	return id, diagnostics
}

func IDStringValue(value int64) types.String {
	return types.StringValue(fmt.Sprintf("%d", value))
}
