package hcclient

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAPIErrorDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name        string
		err         error
		diagnostics diag.Diagnostics
	}{
		{
			name: "hcloud invalid input error",
			err: hcloud.Error{
				Code:    hcloud.ErrorCodeInvalidInput,
				Message: "something is fishy",
				Details: hcloud.ErrorDetailsInvalidInput{
					Fields: []hcloud.ErrorDetailsInvalidInputField{
						{Name: "foobar", Messages: []string{"must be bar", "foo too long"}},
					},
				},
			},
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic(
					"Invalid field in API request",
					`An invalid field was encountered during an API request. The field might not map 1:1 to your terraform resource.

something is fishy (invalid_input)

Field: foobar
Messages:
 - must be bar
 - foo too long
Error code: invalid_input
`),
			},
		},
		{
			name: "hcloud error",
			err: hcloud.Error{
				Code:    hcloud.ErrorCodeRateLimitExceeded,
				Message: "rate limit exceeded",
			},
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic(
					"API request failed",
					`An unexpected error was encountered during an API request.

rate limit exceeded

Error code: rate_limit_exceeded
`),
			},
		},
		{
			name: "generic error",
			err:  fmt.Errorf("something broke :("),
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic(
					"API request failed",
					`An unexpected error was encountered during an API request.

something broke :(
`),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			diags := APIErrorDiagnostics(tc.err)
			if !reflect.DeepEqual(diags, tc.diagnostics) {
				t.Errorf("expected %+v\n\nfound %+v", tc.diagnostics, diags)
			}
		})
	}
}
