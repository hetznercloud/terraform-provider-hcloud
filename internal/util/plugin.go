package util

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ModelFromAPI defines a model than can be read from an API resource.
type ModelFromAPI[API any] interface {
	FromAPI(ctx context.Context, hc API) diag.Diagnostics
}

// ModelToAPI defines a model than can be written to an API resource.
type ModelToAPI[API any] interface {
	ToAPI(ctx context.Context) (API, diag.Diagnostics)
}

// ModelFromTerraform defines a model than can be read from a Terraform type.
type ModelFromTerraform[TF any] interface {
	FromTerraform(ctx context.Context, tf TF) diag.Diagnostics
}

// ModelToTerraform defines a model than can be written to a Terraform type.
type ModelToTerraform[TF any] interface {
	ToTerraform(ctx context.Context) (TF, diag.Diagnostics)
}
