package datasourceutil

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestGetOneResultForLabelSelector(t *testing.T) {
	type item struct{ ID int }

	type args struct {
		resourceName  string
		items         []*item
		labelSelector string
	}
	type testCase struct {
		name       string
		args       args
		want       *item
		diagnostic diag.Diagnostic
	}
	tests := []testCase{
		{
			name: "one item",
			args: args{
				resourceName:  "item",
				items:         []*item{{ID: 1}},
				labelSelector: "foo=bar",
			},
			want:       &item{ID: 1},
			diagnostic: nil,
		},
		{
			name: "zero items",
			args: args{
				resourceName:  "item",
				items:         []*item{},
				labelSelector: "foo=bar",
			},
			want:       nil,
			diagnostic: diag.NewErrorDiagnostic("No item found for label selector", "No item found for label selector.\n\nLabel selector: foo=bar\n"),
		},
		{
			name: "two items",
			args: args{
				resourceName:  "item",
				items:         []*item{{ID: 1}, {ID: 2}},
				labelSelector: "foo=bar",
			},
			want:       nil,
			diagnostic: diag.NewErrorDiagnostic("More than one item found for label selector", "More than one item found for label selector.\n\nLabel selector: foo=bar\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotItem, gotDiagnostic := GetOneResultForLabelSelector(tt.args.resourceName, tt.args.items, tt.args.labelSelector)
			if !reflect.DeepEqual(gotItem, tt.want) {
				t.Errorf("GetOneResultForLabelSelector() got = %v, want %v", gotItem, tt.want)
			}
			if !reflect.DeepEqual(gotDiagnostic, tt.diagnostic) {
				t.Errorf("GetOneResultForLabelSelector() got1 = %v, want %v", gotDiagnostic, tt.diagnostic)
			}
		})
	}
}
