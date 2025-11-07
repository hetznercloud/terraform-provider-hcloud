package validateutil

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestIPValidator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		given types.String
		want  diag.Diagnostics
	}{
		"unknown": {
			given: types.StringUnknown(),
		},
		"null": {
			given: types.StringNull(),
		},
		"valid": {
			given: types.StringValue("10.0.1.2"),
		},
		"invalid": {
			given: types.StringValue("10.0.12"),
			want: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					"Attribute test must be a valid ip, got: 10.0.12",
				),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    testCase.given,
			}
			resp := validator.StringResponse{}

			IPValidator{}.ValidateString(t.Context(), req, &resp)

			assert.Equal(t, testCase.want, resp.Diagnostics)
		})
	}
}
