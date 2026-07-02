package zonerrset_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
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
			TTL:    new(10800),
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
			TTL:    new(600),
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
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
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
					resource.TestCheckResourceAttr(res1.TFID(), "change_protection", "true"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("records"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"value":   knownvalue.StringExact("201.42.91.35"),
								"comment": knownvalue.Null(),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"value":   knownvalue.StringExact("201.42.91.36"),
								"comment": knownvalue.StringExact("some web server"),
							}),
						})),
				},
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
					resource.TestCheckResourceAttr(res1.TFID(), "change_protection", "false"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("records"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"value":   knownvalue.StringExact("42.42.91.35"),
								"comment": knownvalue.Null(),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"value":   knownvalue.StringExact("42.42.91.36"),
								"comment": knownvalue.StringExact("some web server"),
							}),
						})),
				},
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
					resource.TestCheckResourceAttr(res1.TFID(), "change_protection", "false"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(),
						tfjsonpath.New("records"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"value":   knownvalue.StringExact("42.42.91.35"),
								"comment": knownvalue.Null(),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"value":   knownvalue.StringExact("42.42.91.36"),
								"comment": knownvalue.StringExact("some web server"),
							}),
						})),
				},
			},
		},
	})
}

// TestAccZoneRRSetResource_SOA covers https://github.com/hetznercloud/terraform-provider-hcloud/issues/1289.
//
// A SOA RRSet is created by the API together with the parent Zone, so it cannot
// be created a second time. The user must still be able to manage its record
// values: the Create method therefore updates the pre-existing SOA RRSet with
// the configured records in a single apply. The SERIAL field is normalized to 0
// on both sides (the user sets it to 0 in the configuration, the provider
// normalizes it in the state), so plan and apply stay consistent and there is no
// perpetual diff.
func TestAccZoneRRSetResource_SOA(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	// Configured SOA records that differ from the values the API assigns to a
	// freshly created Zone. Before the fix, applying these produced "Provider
	// produced inconsistent result after apply". The SERIAL field is set to 0
	// (see OverrideRecordsSOASerial) as it is managed by the API.
	const soaValue1 = "ns1.example.com. hostmaster.example.com. 0 3600 600 86400 300"
	const soaValue2 = "ns2.example.com. hostmaster.example.com. 0 7200 1800 604800 600"

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
				{Value: soaValue1},
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
				{Value: soaValue2},
			},
		},
	}
	res2.SetRName("soa")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Applying user-provided SOA records must succeed in a single
				// apply (update-on-create), and the state must reflect the
				// configured value.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zonerrset.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "id", "@/SOA"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.#", "1"),
					resource.TestCheckResourceAttr(res1.TFID(), "records.0.value", soaValue1),
				),
			},
			{
				// Re-planning the same configuration produces no diff.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res1,
				),
				PlanOnly: true,
			},
			{
				// Changing the configured SOA records applies the new value.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_rrset", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res2.TFID(), zonerrset.GetAPIResource()),
					resource.TestCheckResourceAttr(res2.TFID(), "id", "@/SOA"),
					resource.TestCheckResourceAttr(res2.TFID(), "records.#", "1"),
					resource.TestCheckResourceAttr(res2.TFID(), "records.0.value", soaValue2),
				),
			},
			{
				// Re-planning the changed configuration produces no diff.
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
