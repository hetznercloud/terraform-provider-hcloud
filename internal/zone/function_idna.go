package zone

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"golang.org/x/net/idna"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
)

func NewIDNAFunction() function.Function {
	return &IDNAFunction{}
}

type IDNAFunction struct{}

func (f *IDNAFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "idna"
}

func (f *IDNAFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Convert a Internationalized Domain Name (IDN) to ASCII",
		MarkdownDescription: `
Converts a Internationalized Domain Name (IDN) to ASCII using Punycode.

The conversion is defined by Golang's IDNA package. See https://pkg.go.dev/golang.org/x/net/idna
for more details.`,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "domain",
				Description: "Domain to convert.",
			},
		},
		Return: function.StringReturn{},
	}

	experimental.DNS.AppendNotice(&resp.Definition.MarkdownDescription)
}

func (f *IDNAFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var domain string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &domain))
	if resp.Error != nil {
		return
	}

	result, err := idna.ToASCII(domain)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("failed to convert domain to ASCII: %s", err))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, result))
}
