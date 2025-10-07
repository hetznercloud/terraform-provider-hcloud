package zonerrset_test

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zonerrset"
)

func TestAccZoneRRSetDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resZone := &zone.RData{
		Zone: schema.Zone{
			Name: fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode: "primary",
		},
	}
	resZone.SetRName("main")

	resZoneRRSet1 := &zonerrset.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name:   "www1",
			Type:   "A",
			Labels: map[string]string{"key": randutil.GenerateID()},
			Records: []schema.ZoneRRSetRecord{
				{Value: "201.42.91.35"},
				{Value: "201.42.91.36", Comment: "some web server"},
			},
		},
	}
	resZoneRRSet1.SetRName("main1")

	resZoneRRSet2 := &zonerrset.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name:   "www2",
			Type:   "A",
			Labels: map[string]string{"key": randutil.GenerateID()},
			Records: []schema.ZoneRRSetRecord{
				{Value: "201.42.91.35"},
				{Value: "201.42.91.36", Comment: "some web server"},
			},
		},
	}
	resZoneRRSet2.SetRName("main2")

	all := &zonerrset.DDataList{
		Zone: resZone.TFID() + ".name",
	}
	all.SetRName("all")

	byLabel := &zonerrset.DDataList{
		Zone:          resZone.TFID() + ".name",
		LabelSelector: labelutil.Selector(resZoneRRSet1.Labels),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", resZoneRRSet1,
					"testdata/r/hcloud_zone_rrset", resZoneRRSet2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", resZoneRRSet1,
					"testdata/r/hcloud_zone_rrset", resZoneRRSet2,
					"testdata/d/hcloud_zone_rrsets", all,
					"testdata/d/hcloud_zone_rrsets", byLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(all.TFID(), "rrsets.#", "4"), // 4 because we include the NS and SOA records

					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.#", "1"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.id", "www1/A"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.name", "www1"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.type", "A"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.labels.key", resZoneRRSet1.Labels["key"]),
					resource.TestCheckNoResourceAttr(byLabel.TFID(), "rrsets.0.ttl"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.records.#", "2"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.records.0.value", "201.42.91.35"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.records.1.value", "201.42.91.36"),
					resource.TestCheckResourceAttr(byLabel.TFID(), "rrsets.0.change_protection", "false"),
				),
			},
		},
	})
}
