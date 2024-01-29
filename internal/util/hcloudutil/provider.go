package hcloudutil

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

func ConfigureClient(providerData any) (*hcloud.Client, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if providerData == nil {
		return nil, diagnostics
	}

	client, ok := providerData.(*hcloud.Client)
	if !ok {
		diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected *hcloud.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil, diagnostics
	}

	return client, diagnostics
}
