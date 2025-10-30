package zonerrset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/zoneutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/experimental"
)

func NewTXTRecordFunction() function.Function {
	return &TXTRecordFunction{}
}

type TXTRecordFunction struct{}

func (f *TXTRecordFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "txt_record"
}

func (f *TXTRecordFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Format a TXT record",
		MarkdownDescription: util.MarkdownDescription(`
Format a TXT record by splitting it in quoted strings of 255 characters.
`),
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "record",
				Description: "Record value to format.",
			},
		},
		Return: function.StringReturn{},
	}

	experimental.DNS.AppendNotice(&resp.Definition.MarkdownDescription)
}

func (f *TXTRecordFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var record string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &record))
	if resp.Error != nil {
		return
	}

	result := zoneutil.FormatTXTRecord(record)

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, result))
}
