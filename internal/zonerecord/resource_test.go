package zonerrset_test

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zonerrset"
)

func TestAccZoneRRSetResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	resZone := &zone.RData{
		Zone: schema.Zone{
			Name: fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode: "primary",
		},
	}
	resZone.SetRName("main")

	res1 := &zonerrset.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name:   "www",
			Type:   "A",
			Labels: map[string]string{"key": "value"},
			TTL:    hcloud.Ptr(10800),
			Records: []schema.ZoneRRSetRecord{
				{Value: "201.42.91.35"},
				{Value: "201.42.91.36", Comment: "some web server"},
			},
		},
		Raw: `change_protection = true`,
	}

	res2 := &zonerrset.RData{
		Zone: res1.Zone,
		ZoneRRSet: schema.ZoneRRSet{
			Name:   res1.Name,
			Type:   res1.Type,
			Labels: map[string]string{"key": "changed"},
			TTL:    hcloud.Ptr(600),
			Records: []schema.ZoneRRSetRecord{
				{Value: "42.42.91.35"},
				{Value: "42.42.91.36", Comment: "some web server"},
			},
		},
		Raw: `change_protection = false`,
	}

	res3 := &zonerrset.RData{
		Zone: res1.Zone,
		ZoneRRSet: schema.ZoneRRSet{
			Name:    res1.Name,
			Type:    res1.Type,
			Records: res2.Records,
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zonerrset.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "id", "www/A"),
					resource.TestCheckResourceAttr(res1.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(res1.TFID(), "labels.key", "value"),
					resource.TestCheckResourceAttr(res1.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.#", "2"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.0.value", "201.42.91.35"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.1.value", "201.42.91.36"),
					resource.TestCheckNoResourceAttr(res1.TFID(), "records.0.comment"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.1.comment", "some web server"),
					resource.TestCheckResourceAttr(res1.TFID(), "change_protection", "true"),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportStateId:     fmt.Sprintf("%s/%s/%s", resZone.Name, res1.Name, res1.Type),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zonerrset.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "id", "www/A"),
					resource.TestCheckResourceAttr(res1.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(res1.TFID(), "labels.key", "changed"),
					resource.TestCheckResourceAttr(res1.TFID(), "ttl", "600"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.#", "2"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.0.value", "42.42.91.35"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.1.value", "42.42.91.36"),
					resource.TestCheckNoResourceAttr(res1.TFID(), "records.0.comment"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.1.comment", "some web server"),
					resource.TestCheckResourceAttr(res1.TFID(), "change_protection", "false"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res3,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zonerrset.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "id", "www/A"),
					resource.TestCheckResourceAttr(res1.TFID(), "name", "www"),
					resource.TestCheckResourceAttr(res1.TFID(), "type", "A"),
					resource.TestCheckResourceAttr(res1.TFID(), "labels.#", "0"),
					resource.TestCheckNoResourceAttr(res1.TFID(), "ttl"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.#", "2"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.0.value", "42.42.91.35"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.1.value", "42.42.91.36"),
					resource.TestCheckNoResourceAttr(res1.TFID(), "records.0.comment"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.1.comment", "some web server"),
					resource.TestCheckResourceAttr(res1.TFID(), "change_protection", "false"),
				),
			},
		},
	})
}
