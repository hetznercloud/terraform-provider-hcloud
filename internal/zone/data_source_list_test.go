package zone_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/labelutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zone"
)

func TestAccZoneDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res1 := &zone.RData{
		Zone: schema.Zone{
			Name:   fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode:   "primary",
			Labels: map[string]string{"key": randutil.GenerateID()},
			TTL:    10800,
		},
	}
	res1.SetRName("main1")

	res2 := &zone.RData{
		Zone: schema.Zone{
			Name:   fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode:   "primary",
			Labels: map[string]string{"key": "value"},
			TTL:    10800,
		},
	}
	res2.SetRName("main2")

	all := &zone.DDataList{}
	all.SetRName("all")

	byLabel := &zone.DDataList{LabelSelector: labelutil.Selector(res1.Labels)}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", res1,
					"testdata/r/hcloud_zone", res2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", res1,
					"testdata/r/hcloud_zone", res2,
					"testdata/d/hcloud_zones", all,
					"testdata/d/hcloud_zones", byLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.#", "1"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.name", res1.Name),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.mode", "primary"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.labels.key", res1.Labels["key"]),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.ttl", "10800"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.primary_nameservers.#", "0"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.delete_protection", "false"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "zones.0.registrar", "other"),
				),
			},
		},
	})
}
