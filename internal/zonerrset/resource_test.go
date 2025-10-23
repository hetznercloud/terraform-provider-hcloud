package zonerrset_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

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

func TestAccZoneRRSetResource_SOA(t *testing.T) {
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
			Name: "@",
			Type: "SOA",
			Records: []schema.ZoneRRSetRecord{
				{Value: "hydrogen.ns.hetzner.com. dns.hetzner.com. 0 86400 10800 3600000 3600"},
			},
		},
	}
	res1.SetRName("soa")

	res2 := &zonerrset.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name: "@",
			Type: "SOA",
			Records: []schema.ZoneRRSetRecord{
				{Value: "hydrogen.ns.hetzner.com. dns.hetzner.com. 0 86400 10800 3600000 600"},
			},
		},
	}
	res2.SetRName("soa")

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
					resource.TestCheckResourceAttr(res1.TFID(), "id", "@/SOA"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.#", "1"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.0.value", "hydrogen.ns.hetzner.com. dns.hetzner.com. 0 86400 10800 3600000 3600"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res2.TFID(), zonerrset.GetAPIResource()),
					resource.TestCheckResourceAttr(res2.TFID(), "id", "@/SOA"),
					resource.TestCheckResourceAttr(res2.TFID(), "records.#", "1"),
					resource.TestCheckResourceAttr(res2.TFID(), "records.0.value", "hydrogen.ns.hetzner.com. dns.hetzner.com. 0 86400 10800 3600000 600"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res2,
				),
				PlanOnly: true,
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res2,
				),
				Destroy: true,
			},
		},
	})
}

func TestOverrideRecordsSOASerial(t *testing.T) {
	testCases := []struct {
		desc        string
		givenType   hcloud.ZoneRRSetType
		givenRecord hcloud.ZoneRRSetRecord
		wantRecord  hcloud.ZoneRRSetRecord
	}{
		{
			desc:        "valid",
			givenType:   hcloud.ZoneRRSetTypeSOA,
			givenRecord: hcloud.ZoneRRSetRecord{Value: `hydrogen.ns.hetzner.com. dns.hetzner.com. 2025102103 86400 10800 3600000 3600`},
			wantRecord:  hcloud.ZoneRRSetRecord{Value: `hydrogen.ns.hetzner.com. dns.hetzner.com. 0 86400 10800 3600000 3600`},
		},
		{
			desc:        "valid minimal",
			givenType:   hcloud.ZoneRRSetTypeSOA,
			givenRecord: hcloud.ZoneRRSetRecord{Value: `hydrogen.ns.hetzner.com. dns.hetzner.com. 2025102103`},
			wantRecord:  hcloud.ZoneRRSetRecord{Value: `hydrogen.ns.hetzner.com. dns.hetzner.com. 0`},
		},
		{
			desc:        "invalid",
			givenType:   hcloud.ZoneRRSetTypeSOA,
			givenRecord: hcloud.ZoneRRSetRecord{Value: `hydrogen.ns.hetzner.com. dns.hetzner.com.`},
			wantRecord:  hcloud.ZoneRRSetRecord{Value: `hydrogen.ns.hetzner.com. dns.hetzner.com.`},
		},
		{
			desc:        "skipped",
			givenType:   hcloud.ZoneRRSetTypeTXT,
			givenRecord: hcloud.ZoneRRSetRecord{Value: `"one" "two" "three" "four"`},
			wantRecord:  hcloud.ZoneRRSetRecord{Value: `"one" "two" "three" "four"`},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			rrset := &hcloud.ZoneRRSet{
				Type:    testCase.givenType,
				Records: []hcloud.ZoneRRSetRecord{testCase.givenRecord},
			}
			zonerrset.OverrideRecordsSOASerial(rrset)

			require.Equal(t, testCase.wantRecord, rrset.Records[0])
		})
	}
}
