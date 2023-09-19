package hcclient

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// APIErrorDiagnostics creates diagnostics from the errors that occurred during an API requests.
func APIErrorDiagnostics(err error) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	var hcloudErr hcloud.Error

	if errors.As(err, &hcloudErr) {
		if hcloud.IsError(hcloudErr, hcloud.ErrorCodeInvalidInput) {
			invalidInput := hcloudErr.Details.(hcloud.ErrorDetailsInvalidInput)
			for _, field := range invalidInput.Fields {
				diagnostics.AddError(
					"Invalid field in API request",
					fmt.Sprintf(
						"An invalid field was encountered during an API request. "+
							"The field might not map 1:1 to your terraform resource.\n\n"+
							"%s\n\n"+
							"Field: %s\n"+
							"Messages: %s\n"+
							"Error code: %s\n",
						err.Error(), field.Name, field.Messages, hcloudErr.Code,
					))
			}
			return diagnostics
		}

		diagnostics.AddError(
			"API request failed",
			fmt.Sprintf(
				"An unexpected error was encountered during an API request.\n\n"+
					"%s\n\n"+
					"Error code: %s\n",
				hcloudErr.Message, hcloudErr.Code,
			),
		)
		return diagnostics
	}

	diagnostics.AddError(
		"API request failed",
		fmt.Sprintf(
			"An unexpected error was encountered during an API request.\n\n"+
				"%s\n",
			err.Error(),
		),
	)
	return diagnostics
}
