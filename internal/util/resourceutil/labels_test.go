package resourceutil

import (
	"context"
	"reflect"
	"sort"
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

			// Diagnostics might be unordered, which add flakiness to the diff below.
			sort.Slice(response.Diagnostics, func(i, j int) bool {
				iPath := response.Diagnostics[i].(diag.DiagnosticWithPath).Path()
				jPath := response.Diagnostics[j].(diag.DiagnosticWithPath).Path()
				return iPath.String() < jPath.String()
			})

			if diff := cmp.Diff(response.Diagnostics, test.expected); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestLabelsMapValueFrom(t *testing.T) {
	tests := []struct {
		name  string
		in    map[string]string
		want  types.Map
		diags diag.Diagnostics
	}{
		{
			name:  "Map with Labels",
			in:    map[string]string{"foo": "bar"},
			want:  types.MapValueMust(types.StringType, map[string]attr.Value{"foo": types.StringValue("bar")}),
			diags: nil,
		},
		{
			name:  "Empty Map",
			in:    map[string]string{},
			want:  types.MapNull(types.StringType),
			diags: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels, diags := LabelsMapValueFrom(context.Background(), tt.in)
			if !reflect.DeepEqual(labels, tt.want) {
				t.Errorf("LabelsMapValueFrom() got = %v, want %v", labels, tt.want)
			}
			if !reflect.DeepEqual(diags, tt.diags) {
				t.Errorf("LabelsMapValueFrom() got1 = %v, want %v", diags, tt.diags)
			}
		})
	}
}
