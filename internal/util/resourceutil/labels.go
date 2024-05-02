package resourceutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

var _ validator.Map = (*labelsValidator)(nil)

type labelsValidator struct{}

func (l labelsValidator) Description(_ context.Context) string {
	return "labels must conform to the labels format: https://docs.hetzner.cloud/#labels"
}

func (l labelsValidator) MarkdownDescription(_ context.Context) string {
	return "labels must conform to the [labels format](https://docs.hetzner.cloud/#labels)"
}

func (l labelsValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	for k, v := range req.ConfigValue.Elements() {
		label := map[string]interface{}{k: v.(types.String).ValueString()}

		if ok, err := hcloud.ValidateResourceLabels(label); !ok {
			resp.Diagnostics.AddAttributeError(req.Path.AtMapKey(k), "Invalid label", err.Error())
		}
	}
}

// LabelsSchema returns a map attribute schema with validation for the labels field shared by multiple resources.
func LabelsSchema() schema.MapAttribute {
	return schema.MapAttribute{
		MarkdownDescription: "User-defined [labels](https://docs.hetzner.cloud/#labels) (key-value pairs) for the resource.",
		Optional:            true,
		Computed:            true, // Required to use Default
		ElementType:         types.StringType,
		// In the resource schemas, labels can be null, but the API always returns an empty object for labels.
		// To avoid a data consistency issue, we set the default value to the empty object.
		Default: mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
		Validators: []validator.Map{
			labelsValidator{},
		},
	}
}

// LabelsMapValueFrom prepare the labels from the API to be assigned into the resource model.
func LabelsMapValueFrom(ctx context.Context, in map[string]string) (types.Map, diag.Diagnostics) {
	return types.MapValueFrom(ctx, types.StringType, in)
}
