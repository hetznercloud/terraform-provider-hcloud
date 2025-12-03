package zonerecord_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zone"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zonerecord"
)

func TestAccZoneRecordResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resZone := &zone.RData{
		Zone: schema.Zone{
			Name: fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode: "primary",
		},
	}
	resZone.SetRName("main")

	res1 := &zonerecord.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name: "www",
			Type: "A",
			TTL:  hcloud.Ptr(10800),
			Records: []schema.ZoneRRSetRecord{
				{Value: "201.42.91.36", Comment: "some web server"},
			},
		},
	}
	res1.SetRName("record")

	res2 := &zonerecord.RData{
		Zone: res1.Zone,
		ZoneRRSet: schema.ZoneRRSet{
			Name: res1.Name,
			Type: res1.Type,
			TTL:  hcloud.Ptr(600),
			Records: []schema.ZoneRRSetRecord{
				{Value: "42.42.91.35"},
			},
		},
	}
	res2.SetRName("record")

	res3 := &zonerecord.RData{
		Zone: res1.Zone,
		ZoneRRSet: schema.ZoneRRSet{
			Name:    res1.Name,
			Type:    res1.Type,
			Records: res2.Records,
		},
	}
	res3.SetRName("record")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_record", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zonerecord.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(res1.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(res1.TFID(), "value", "201.42.91.36"),
					resource.TestCheckResourceAttr(res1.TFID(), "comment", "some web server"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_record", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res2.TFID(), zonerecord.GetAPIResource()),
					resource.TestCheckResourceAttr(res2.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(res2.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(res2.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(res2.TFID(), "value", "42.42.91.35"),
					resource.TestCheckNoResourceAttr(res2.TFID(), "comment"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_record", res3,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res3.TFID(), zonerecord.GetAPIResource()),
					resource.TestCheckResourceAttr(res3.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(res3.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(res3.TFID(), "ttl", ""),
					resource.TestCheckResourceAttr(res3.TFID(), "value", "42.42.91.35"),
					resource.TestCheckNoResourceAttr(res3.TFID(), "comment"),
				),
			},
		},
	})
}
