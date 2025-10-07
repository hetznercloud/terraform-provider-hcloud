package zonerrset_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/labelutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zone"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zonerrset"
)

func TestAccZoneRRSetDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resZone := &zone.RData{
		Zone: schema.Zone{
			Name: fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode: "primary",
		},
	}
	resZone.SetRName("main")

	resZoneRRSet := &zonerrset.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name:   "www",
			Type:   "A",
			Labels: map[string]string{"key": randutil.GenerateID()},
			TTL:    hcloud.Ptr(10800),
			Records: []schema.ZoneRRSetRecord{
				{Value: "201.42.91.35"},
				{Value: "201.42.91.36", Comment: "some web server"},
			},
		},
	}
	resZoneRRSet.SetRName("main")

	byID := &zonerrset.DData{
		Zone: resZoneRRSet.Zone,
		ID:   resZoneRRSet.TFID() + ".id",
	}
	byID.SetRName("by_id")
	byNameAndType := &zonerrset.DData{
		Zone: resZoneRRSet.Zone,
		Name: resZoneRRSet.TFID() + ".name",
		Type: resZoneRRSet.TFID() + ".type",
	}
	byNameAndType.SetRName("by_name_and_type")
	byLabel := &zonerrset.DData{
		Zone:          resZoneRRSet.Zone,
		LabelSelector: labelutil.Selector(resZoneRRSet.Labels),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", resZoneRRSet,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", resZoneRRSet,
					"testdata/d/hcloud_zone_rrset", byNameAndType,
					"testdata/d/hcloud_zone_rrset", byID,
					"testdata/d/hcloud_zone_rrset", byLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "zone", resZone.Name),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "id", "www/A"),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "labels.key", resZoneRRSet.Labels["key"]),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "records.#", "2"),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "records.0.value", "201.42.91.35"),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "records.1.value", "201.42.91.36"),
					resource.TestCheckResourceAttr(byNameAndType.TFID(), "change_protection", "false"),

					resource.TestCheckResourceAttr(byID.TFID(), "zone", resZone.Name),
					resource.TestCheckResourceAttr(byID.TFID(), "id", "www/A"),
					resource.TestCheckResourceAttr(byID.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(byID.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(byID.TFID(), "labels.key", resZoneRRSet.Labels["key"]),
					resource.TestCheckResourceAttr(byID.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(byID.TFID(), "records.#", "2"),
					resource.TestCheckResourceAttr(byID.TFID(), "records.0.value", "201.42.91.35"),
					resource.TestCheckResourceAttr(byID.TFID(), "records.1.value", "201.42.91.36"),
					resource.TestCheckResourceAttr(byID.TFID(), "change_protection", "false"),

					resource.TestCheckResourceAttr(byLabel.TFID(), "zone", resZone.Name),
					resource.TestCheckResourceAttr(byLabel.TFID(), "id", "www/A"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "labels.key", resZoneRRSet.Labels["key"]),
					resource.TestCheckResourceAttr(byLabel.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "records.#", "2"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "records.0.value", "201.42.91.35"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "records.1.value", "201.42.91.36"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "change_protection", "false"),
				),
			},
		},
	})
}
