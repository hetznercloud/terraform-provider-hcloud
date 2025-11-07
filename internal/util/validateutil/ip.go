package validateutil

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = IPValidator{}

type IPValidator struct{}

func IP() IPValidator {
	return IPValidator{}
}

func (v IPValidator) Description(_ context.Context) string {
	return "must be a valid ip"
}

func (v IPValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v IPValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	raw := req.ConfigValue.ValueString()
	ip := net.ParseIP(raw)
	if ip == nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(req.Path, v.Description(ctx), raw))
	}
}
