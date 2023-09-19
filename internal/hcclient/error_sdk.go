package hcclient

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

// ErrorToDiag creates a terraform diag
// When some hcloud errors are passed it enriches the default
// Error() function from them with a few more details to make
// them more understandable for users
func ErrorToDiag(err error) diag.Diagnostics {
	var hcloudErr hcloud.Error

	if !errors.As(err, &hcloudErr) {
		return diag.FromErr(err)
	}
	if hcloud.IsError(err, hcloud.ErrorCodeInvalidInput) {
		return enrichInvalidInput(hcloudErr)
	}
	return diag.FromErr(err)
}
func enrichInvalidInput(err hcloud.Error) diag.Diagnostics {
	ie := err.Details.(hcloud.ErrorDetailsInvalidInput)
	invalidInputs := make([]string, len(ie.Fields))
	for i, v := range ie.Fields {
		invalidInputs[i] = fmt.Sprintf("%s => %s", v.Name, v.Messages)
	}
	return diag.Errorf("%s: %s", err.Error(), invalidInputs)
}
