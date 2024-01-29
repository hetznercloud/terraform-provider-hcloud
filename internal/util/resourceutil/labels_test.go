package resourceutil

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestLabelsValidator_ValidateMap(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val      types.Map
		expected diag.Diagnostics
	}
	tests := map[string]testCase{
		"invalid": {
			val: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"key1":      types.StringValue("ö-invalid"),
					"ö-invalid": types.StringValue("second"),
				},
			),
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("key1"),
					"Invalid label",
					"label value 'ö-invalid' (key: key1) is not correctly formatted",
				),
				diag.NewAttributeErrorDiagnostic(
					path.Root("test").AtMapKey("ö-invalid"),
					"Invalid label",
					"label key 'ö-invalid' is not correctly formatted",
				),
			},
		},
		"valid": {
			val: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"key1": types.StringValue("first"),
					"key2": types.StringValue("second"),
				},
			),
			expected: diag.Diagnostics{},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := validator.MapRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.MapResponse{}
			labelsValidator{}.ValidateMap(context.Background(), request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expected); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
