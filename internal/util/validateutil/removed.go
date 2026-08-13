package validateutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = IsRemovedValidator{}

type IsRemovedValidator struct {
	details string
}

func IsRemoved(details string) IsRemovedValidator {
	return IsRemovedValidator{details: details}
}

func (v IsRemovedValidator) Description(_ context.Context) string { return "Attribute was removed" }

func (v IsRemovedValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v IsRemovedValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
		req.Path,
		"Removed Attribute",
		v.details,
	))
}
