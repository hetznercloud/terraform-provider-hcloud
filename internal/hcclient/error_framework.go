package hcclient

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

func APIErrorDiagnostic(resource string, err error) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	var hcloudErr hcloud.Error

	if errors.As(err, &hcloudErr) {
		if hcloud.IsError(hcloudErr, hcloud.ErrorCodeInvalidInput) {
			invalidInput := hcloudErr.Details.(hcloud.ErrorDetailsInvalidInput)
			for _, field := range invalidInput.Fields {
				diagnostics.AddError(
					"Invalid field in API request",
					fmt.Sprintf(
						`An invalid field was encountered while doing the request to %s. The field might not map 1:1 to your terraform resource.

%s => %s

Error code: %s
`,
						resource, field.Name, field.Messages, hcloudErr.Code,
					))
			}
			return diagnostics
		}

		diagnostics.AddError(
			fmt.Sprintf("Request to %s failed", resource),
			fmt.Sprintf(
				`An unexpected error was encountered while doing the request to %s.
%s

Error code: %s`,
				resource, hcloudErr.Message, hcloudErr.Code,
			),
		)
		return diagnostics
	}

	diagnostics.AddError(
		fmt.Sprintf("Request to %s failed", resource),
		fmt.Sprintf(
			"An unexpected error was encountered while doing the request to %s.\n\n%s\n",
			resource, err.Error(),
		),
	)
	return diagnostics
}
