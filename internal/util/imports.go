package util

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func InvalidImportID(expected string, id string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Invalid import ID",
		fmt.Sprintf("Received an invalid import ID, expected '%s' but received '%s'.", expected, id),
	)
}
