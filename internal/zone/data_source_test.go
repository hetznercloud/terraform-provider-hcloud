package zone_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/labelutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zone"
)

func TestAccZoneDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &zone.RData{
		Zone: schema.Zone{
			Name:   fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode:   "primary",
			Labels: map[string]string{"key": randutil.GenerateID()},
			TTL:    10800,
		},
	}
	res.SetRName("main")

	byID := &zone.DData{
		ID: res.TFID() + ".id",
	}
	byID.SetRName("by_id")
	byName := &zone.DData{
		Name: res.TFID() + ".name",
	}
	byName.SetRName("by_name")
	byLabel := &zone.DData{
		LabelSelector: labelutil.Selector(res.Labels),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", res,
					"testdata/d/hcloud_zone", byName,
					"testdata/d/hcloud_zone", byID,
					"testdata/d/hcloud_zone", byLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(byName.TFID(), "name", res.Name),
					resource.TestCheckResourceAttr(byName.TFID(), "mode", "primary"),
					resource.TestCheckResourceAttr(byName.TFID(), "labels.key", res.Labels["key"]),
					resource.TestCheckResourceAttr(byName.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(byName.TFID(), "primary_nameservers.#", "0"),
					resource.TestCheckResourceAttr(byName.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(byName.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(byName.TFID(), "registrar", "other"),

					resource.TestCheckResourceAttr(byID.TFID(), "name", res.Name),
					resource.TestCheckResourceAttr(byID.TFID(), "mode", "primary"),
					resource.TestCheckResourceAttr(byID.TFID(), "labels.key", res.Labels["key"]),
					resource.TestCheckResourceAttr(byID.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(byID.TFID(), "primary_nameservers.#", "0"),
					resource.TestCheckResourceAttr(byID.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(byID.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(byID.TFID(), "registrar", "other"),

					resource.TestCheckResourceAttr(byLabel.TFID(), "name", res.Name),
					resource.TestCheckResourceAttr(byLabel.TFID(), "mode", "primary"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "labels.key", res.Labels["key"]),
					resource.TestCheckResourceAttr(byLabel.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "primary_nameservers.#", "0"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "registrar", "other"),
				),
			},
		},
	})
}
