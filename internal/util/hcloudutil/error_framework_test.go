package hcloudutil

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

func TestAPIErrorDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name          string
		err           error
		errRaw        map[string]interface{}
		errStatusCode int
		diagnostics   diag.Diagnostics
	}{
		{
			name: "hcloud invalid input error",
			errRaw: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "invalid_input",
					"message": "something is fishy",
					"details": map[string]interface{}{
						"fields": []map[string]interface{}{
							{"name": "foobar", "messages": []string{"must be bar", "foo too long"}},
						},
					},
				},
			},
			errStatusCode: http.StatusBadRequest,
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
Status code: 400
`),
			},
		},
		{
			name: "hcloud error",
			err: hcloud.Error{
				Code:    hcloud.ErrorCodeRateLimitExceeded,
				Message: "rate limit exceeded",
			},
			errRaw: map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "rate_limit_exceeded",
					"message": "rate limit exceeded",
				},
			},
			errStatusCode: http.StatusTooManyRequests,
			diagnostics: []diag.Diagnostic{
				diag.NewErrorDiagnostic(
					"API request failed",
					`An unexpected error was encountered during an API request.

rate limit exceeded

Error code: rate_limit_exceeded
Status code: 429
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
			err := tc.err
			if tc.errRaw != nil {
				err = hcloudErrorFromErrorAndStatus(tc.errRaw, tc.errStatusCode)
			}

			diags := APIErrorDiagnostics(err)
			if !reflect.DeepEqual(diags, tc.diagnostics) {
				t.Errorf("expected %+v\n\nfound %+v", tc.diagnostics, diags)
			}
		})
	}
}

// hcloudErrorFromErrorAndStatus is a hack to fill the private field `hcloud.Error.response` with a sensible response
// so we can fully test the logged error messages.
func hcloudErrorFromErrorAndStatus(errRaw map[string]interface{}, statusCode int) error {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client := hcloud.NewClient(
		hcloud.WithEndpoint(server.URL),
		hcloud.WithToken("token"),
	)

	mux.HandleFunc("/actions/1", func(w http.ResponseWriter, r *http.Request) { // nolint:revive
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		err := json.NewEncoder(w).Encode(errRaw)
		if err != nil {
			log.Fatal(err)
		}
	})

	_, _, err := client.Action.GetByID(context.Background(), 1)
	return err
}

func TestNotFoundDiagnostics(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		actual   diag.Diagnostic
		expected diag.Diagnostic
	}{
		{
			name:     "location with name",
			actual:   NotFoundDiagnostic("location", "name", "my-location"),
			expected: diag.NewErrorDiagnostic("Resource not found", "Resource (location) was not found: name=my-location"),
		},
		{
			name:     "location with id",
			actual:   NotFoundDiagnostic("location", "id", 123),
			expected: diag.NewErrorDiagnostic("Resource not found", "Resource (location) was not found: id=123"),
		},
		{
			name:     "ssh key with name",
			actual:   NotFoundDiagnostic("ssh_key", "name", "my-ssh-key"),
			expected: diag.NewErrorDiagnostic("Resource not found", "Resource (ssh_key) was not found: name=my-ssh-key"),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, testCase.actual)
		})
	}
}
