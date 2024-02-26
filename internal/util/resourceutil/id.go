package resourceutil

import (
	"fmt"
	"math"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ParseID(value types.String) (int, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	id, err := strconv.ParseInt(value.ValueString(), 10, 64)
	if err != nil {
		diagnostics.AddAttributeError(
			path.Root("id"),
			"Could not parse resource ID",
			err.Error(),
		)
		return 0, diagnostics
	}
	if id > math.MaxInt32 && strconv.IntSize == 32 {
		diagnostics.AddAttributeError(
			path.Root("id"),
			"ID is larger than your system supports.",
			"The current version of the provider has problems with IDs > 32bit on 32 bit systems. If possible, switch to a 64 bit system for now. See https://github.com/hetznercloud/hcloud-go/issues/263")
	}
	return int(id), diagnostics
}

func IDStringValue(value int) types.String {
	return types.StringValue(fmt.Sprintf("%d", value))
}
