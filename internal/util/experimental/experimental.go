package experimental

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Experimental holds the details about an experimental product.
//
// Usage:
//
//	var Product = New("Product name", "https://docs.hetzner.cloud/changelog#new-product")
//
//	func (r *Resource) Configure(_ context.Context, _ resource.ConfigureRequest, resp *resource.ConfigureResponse) {
//		experimental.Product.AppendDiagnostic(&resp.Diagnostics)
//	}
//
//	func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
//		resp.Schema.MarkdownDescription = `
//	Manage a Hetzner Cloud Product.
//
//	See https://docs.hetzner.cloud/reference/cloud#product for more details.
//	`
//		experimental.Product.AppendNotice(&resp.Schema.MarkdownDescription)
//	}
type Experimental struct {
	product string
	url     string

	diagnostic diag.Diagnostic
	notice     string
}

func New(product string, url string) Experimental {
	e := Experimental{product: product, url: url}

	e.diagnostic = diag.NewWarningDiagnostic(
		fmt.Sprintf("Experimental: %s", e.product),
		fmt.Sprintf(`%s is experimental, breaking changes may occur within minor releases.

See %s for more details.
`, e.product, e.url))

	e.notice = fmt.Sprintf(`

**Experimental:** %s is experimental, breaking changes may occur within minor releases.
See %s for more details.
`, e.product, e.url)

	return e
}

// AppendDiagnostic adds a warning about the experimental product to the provided diagnostics.
func (e Experimental) AppendDiagnostic(diags *diag.Diagnostics) {
	diags.Append(e.diagnostic)
}

// AppendNotice adds a warning about the experimental product to the provided description.
func (e Experimental) AppendNotice(description *string) {
	*description = strings.TrimSuffix(*description, "\n") + e.notice
}
