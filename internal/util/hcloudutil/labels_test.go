package hcloudutil

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestTerraformLabelsToHCloud(t *testing.T) {
	tests := []struct {
		name            string
		inputLabels     types.Map
		wantLabels      *map[string]string
		wantDiagnostics bool
	}{
		{
			name:            "Some Labels",
			inputLabels:     types.MapValueMust(types.StringType, map[string]attr.Value{"key1": types.StringValue("value1")}),
			wantLabels:      &map[string]string{"key1": "value1"},
			wantDiagnostics: false,
		},
		{
			name:            "Empty Labels",
			inputLabels:     types.MapNull(types.StringType),
			wantLabels:      &map[string]string{},
			wantDiagnostics: false,
		},
		{
			name:            "Invalid Map Labels",
			inputLabels:     types.MapValueMust(types.BoolType, map[string]attr.Value{"key1": types.BoolValue(true)}),
			wantLabels:      &map[string]string{},
			wantDiagnostics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outputLabels map[string]string
			diagnostics := TerraformLabelsToHCloud(context.Background(), tt.inputLabels, &outputLabels)
			assert.Equalf(t, tt.wantDiagnostics, diagnostics != nil, "Unexpected Diagnostics: %v", diagnostics)
			assert.Equalf(t, *tt.wantLabels, outputLabels, "Unexpected Labels")
		})
	}
}
