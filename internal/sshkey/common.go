package sshkey

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/resourceutil"
)

type resourceData struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	PublicKey   types.String `tfsdk:"public_key"`
	Labels      types.Map    `tfsdk:"labels"`
}

func populateResourceData(ctx context.Context, data *resourceData, in *hcloud.SSHKey) diag.Diagnostics {
	var diags diag.Diagnostics
	var newDiags diag.Diagnostics

	data.ID = resourceutil.IDStringValue(in.ID)
	data.Name = types.StringValue(in.Name)
	data.Fingerprint = types.StringValue(in.Fingerprint)
	data.PublicKey = types.StringValue(in.PublicKey)

	data.Labels, newDiags = resourceutil.LabelsMapValueFrom(ctx, in.Labels)
	diags.Append(newDiags...)

	return diags
}
