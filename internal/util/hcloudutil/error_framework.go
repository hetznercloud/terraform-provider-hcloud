package hcloudutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

// APIErrorDiagnostics creates diagnostics from the errors that occurred during an API requests.
func APIErrorDiagnostics(err error) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	var hcloudErr hcloud.Error

	if errors.As(err, &hcloudErr) {
		statusCodeMessage := ""
		if errResponse := hcloudErr.Response(); errResponse != nil {
			statusCodeMessage = fmt.Sprintf("Status code: %d\n", errResponse.StatusCode)
		}

		if hcloud.IsError(hcloudErr, hcloud.ErrorCodeInvalidInput) {
			invalidInput := hcloudErr.Details.(hcloud.ErrorDetailsInvalidInput)
			for _, field := range invalidInput.Fields {
				messages := make([]string, 0, len(field.Messages))
				for _, message := range field.Messages {
					messages = append(messages, fmt.Sprintf(" - %s", message))
				}

				diagnostics.AddError(
					"Invalid field in API request",
					fmt.Sprintf(
						"An invalid field was encountered during an API request. "+
							"The field might not map 1:1 to your terraform resource.\n\n"+
							"%s\n\n"+
							"Field: %s\n"+
							"Messages:\n%s\n"+
							"Error code: %s\n"+
							"%s",
						err.Error(), field.Name, strings.Join(messages, "\n"), hcloudErr.Code, statusCodeMessage,
					))
			}
			return diagnostics
		}

		diagnostics.AddError(
			"API request failed",
			fmt.Sprintf(
				"An unexpected error was encountered during an API request.\n\n"+
					"%s\n\n"+
					"Error code: %s\n"+
					"%s",
				hcloudErr.Message, hcloudErr.Code, statusCodeMessage,
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

// APIErrorIsNotFound check whether the error is an API request Not Found error.
func APIErrorIsNotFound(err error) bool {
	var hcloudErr hcloud.Error
	if errors.As(err, &hcloudErr) {
		return hcloud.IsError(hcloudErr, hcloud.ErrorCodeNotFound)
	}
	return false
}
