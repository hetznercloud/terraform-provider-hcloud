package testmux

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"hcloud": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			providerFactory, err := tfhcloud.GetMuxedProvider(ctx)
			if err != nil {
				return nil, err
			}

			return providerFactory(), nil
		},
	}
}
