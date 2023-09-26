package e2etests

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hetznercloud/hcloud-go/hcloud"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
)

const (
	// TestImage is the system image that is used in all tests
	TestImage = "ubuntu-20.04"

	// TestServerType is the default server type used in all tests
	TestServerType = "cx11"

	// TestArchitecture is the default architecture used in all tests, should match the architecture of the TestServerType.
	TestArchitecture = hcloud.ArchitectureX86

	// TestLoadBalancerType is the default Load Balancer type used in all tests
	TestLoadBalancerType = "lb11"

	// TestDataCenter is the default datacenter where we execute our tests.
	TestDataCenter = "hel1-dc2"

	// TestLocationName is the default location where we execute our tests.
	TestLocationName = "hel1"
)

func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"hcloud": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()

			upgradedSdkServer, err := tf5to6server.UpgradeServer(
				ctx,
				tfhcloud.Provider().GRPCProvider,
			)

			if err != nil {
				return nil, err
			}

			providers := []func() tfprotov6.ProviderServer{
				providerserver.NewProtocol6(tfhcloud.NewPluginProvider()),
				func() tfprotov6.ProviderServer {
					return upgradedSdkServer
				},
			}

			muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

			if err != nil {
				return nil, err
			}

			return muxServer.ProviderServer(), nil
		},
	}
}

// PreCheck checks if all conditions for an acceptance test are
// met.
func PreCheck(t *testing.T) func() {
	return func() {
		if v := os.Getenv("HCLOUD_TOKEN"); v == "" {
			t.Fatal("HCLOUD_TOKEN must be set for acceptance tests")
		}
	}
}
