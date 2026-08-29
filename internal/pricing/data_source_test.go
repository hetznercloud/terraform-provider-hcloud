package pricing_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/pricing"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccPricingDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &pricing.DData{}
	res.SetRName("test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_pricing", res,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(res.TFID(), "currency"),
					resource.TestCheckResourceAttrSet(res.TFID(), "vat_rate"),
					resource.TestCheckResourceAttrSet(res.TFID(), "image.per_gb_month.net"),
					resource.TestCheckResourceAttrSet(res.TFID(), "server_backup.percentage"),
					resource.TestCheckResourceAttrSet(res.TFID(), "volume.per_gb_month.net"),
					resource.TestCheckResourceAttrSet(res.TFID(), "floating_ips.0.type"),
					resource.TestCheckResourceAttrSet(res.TFID(), "floating_ips.0.prices.0.location"),
					resource.TestCheckResourceAttrSet(res.TFID(), "primary_ips.0.type"),
					resource.TestCheckResourceAttrSet(res.TFID(), "primary_ips.0.prices.0.hourly.net"),
					resource.TestCheckResourceAttrSet(res.TFID(), "server_types.0.name"),
					resource.TestCheckResourceAttrSet(res.TFID(), "server_types.0.prices.0.monthly.net"),
					resource.TestCheckResourceAttrSet(res.TFID(), "server_types.0.prices.0.included_traffic"),
					resource.TestCheckResourceAttrSet(res.TFID(), "load_balancer_types.0.name"),
					resource.TestCheckResourceAttrSet(res.TFID(), "load_balancer_types.0.prices.0.monthly.net"),
				),
			},
		},
	})
}
