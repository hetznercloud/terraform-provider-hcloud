package zonerecord_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

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

	resA1 := &zonerecord.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name: "www",
			Type: "A",
			Records: []schema.ZoneRRSetRecord{
				{Value: "201.42.91.36", Comment: "some web server"},
			},
		},
	}
	resA1.SetRName("record_a")

	resB1 := &zonerecord.RData{
		Zone: resZone.TFID() + ".name",
		ZoneRRSet: schema.ZoneRRSet{
			Name: "www",
			Type: "A",
			Records: []schema.ZoneRRSetRecord{
				{Value: "201.42.91.37", Comment: "some other server"},
			},
		},
	}
	resB1.SetRName("record_b")

	resA2 := testtemplate.DeepCopy(t, resA1)
	resA2.Records[0].Value = "42.42.91.35"
	resA2.Records[0].Comment = ""

	resB2 := testtemplate.DeepCopy(t, resB1)
	resB2.Records[0].Comment = "updated comment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_record", resA1,
					"testdata/r/hcloud_zone_record", resB1,
				),

				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact("www")),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("value"), knownvalue.StringExact("201.42.91.36")),
					statecheck.ExpectKnownValue(resA1.TFID(), tfjsonpath.New("comment"), knownvalue.StringExact("some web server")),

					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact("www")),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("value"), knownvalue.StringExact("201.42.91.37")),
					statecheck.ExpectKnownValue(resB1.TFID(), tfjsonpath.New("comment"), knownvalue.StringExact("some other server")),
				},
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_zone", resZone,
					"testdata/r/hcloud_zone_record", resA2,
					"testdata/r/hcloud_zone_record", resB2,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resA2.TFID(), tfjsonpath.New("name"), knownvalue.StringExact("www")),
					statecheck.ExpectKnownValue(resA2.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resA2.TFID(), tfjsonpath.New("value"), knownvalue.StringExact("42.42.91.35")),
					statecheck.ExpectKnownValue(resA2.TFID(), tfjsonpath.New("comment"), knownvalue.StringExact("")),

					statecheck.ExpectKnownValue(resB2.TFID(), tfjsonpath.New("name"), knownvalue.StringExact("www")),
					statecheck.ExpectKnownValue(resB2.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue(resB2.TFID(), tfjsonpath.New("value"), knownvalue.StringExact("201.42.91.37")),
					statecheck.ExpectKnownValue(resB2.TFID(), tfjsonpath.New("comment"), knownvalue.StringExact("updated comment")),
				},
			},
		},
	})
}
