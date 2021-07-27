module github.com/hetznercloud/terraform-provider-hcloud

require (
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
	github.com/hetznercloud/hcloud-go v1.29.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/net v0.0.0-20210326060303-6b1517762897
)

go 1.16

replace github.com/hetznercloud/hcloud-go => hetzner.cloud/integrations/hcloud-go v1.29.0-rc.5
