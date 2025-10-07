package zone_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/zone"
)

func TestAccZoneResource_Primary(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res1 := &zone.RData{
		Zone: schema.Zone{
			Name:   fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode:   "primary",
			Labels: map[string]string{"key": "value"},
			TTL:    10800,
		},
		Raw: `delete_protection = true`,
	}

	res2 := &zone.RData{
		Zone: schema.Zone{
			Name:   res1.Name,
			Mode:   res1.Mode,
			Labels: map[string]string{"key": "changed"},
			TTL:    600,
		},
		Raw: `delete_protection = false`,
	}

	res3 := &zone.RData{
		Zone: schema.Zone{
			Name: res1.Name,
			Mode: res1.Mode,
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_zone", res1),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zone.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "name", res1.Name),
					resource.TestCheckResourceAttr(res1.TFID(), "mode", "primary"),
					resource.TestCheckResourceAttr(res1.TFID(), "labels.key", "value"),
					resource.TestCheckResourceAttr(res1.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.#", "0"),
					resource.TestCheckResourceAttr(res1.TFID(), "delete_protection", "true"),
					resource.TestCheckResourceAttr(res1.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(res1.TFID(), "registrar", "other"),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportStateId:     res1.Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_zone", res2),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zone.GetAPIResource()),
					resource.TestCheckResourceAttr(res2.TFID(), "name", res2.Name),
					resource.TestCheckResourceAttr(res2.TFID(), "mode", "primary"),
					resource.TestCheckResourceAttr(res2.TFID(), "labels.key", "changed"),
					resource.TestCheckResourceAttr(res2.TFID(), "ttl", "600"),
					resource.TestCheckResourceAttr(res2.TFID(), "primary_nameservers.#", "0"),
					resource.TestCheckResourceAttr(res2.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(res1.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(res1.TFID(), "registrar", "other"),
				),
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_zone", res3),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zone.GetAPIResource()),
					resource.TestCheckResourceAttr(res3.TFID(), "name", res3.Name),
					resource.TestCheckResourceAttr(res3.TFID(), "mode", "primary"),
					resource.TestCheckResourceAttr(res3.TFID(), "labels.#", "0"),
					resource.TestCheckResourceAttr(res3.TFID(), "ttl", "3600"),
					resource.TestCheckResourceAttr(res3.TFID(), "primary_nameservers.#", "0"),
					resource.TestCheckResourceAttr(res3.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(res1.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(res1.TFID(), "registrar", "other"),
				),
			},
		},
	})
}

func TestAccZoneResource_Secondary(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res1 := &zone.RData{
		Zone: schema.Zone{
			Name: fmt.Sprintf("example-%s.com", randutil.GenerateID()),
			Mode: "secondary",
			PrimaryNameservers: []schema.ZonePrimaryNameserver{
				{Address: "201.34.67.42"},
				{Address: "201.34.67.43", Port: 5353},
			},
		},
	}

	res2 := &zone.RData{
		Zone: schema.Zone{
			Name:   res1.Name,
			Mode:   res1.Mode,
			Labels: map[string]string{"key": "value"},
			TTL:    10800,
			PrimaryNameservers: []schema.ZonePrimaryNameserver{
				{Address: "89.34.67.42"},
				{Address: "89.34.67.43", Port: 5353},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(zone.ResourceType, zone.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_zone", res1),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zone.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "name", res1.Name),
					resource.TestCheckResourceAttr(res1.TFID(), "mode", "secondary"),
					resource.TestCheckResourceAttr(res1.TFID(), "labels.#", "0"),
					resource.TestCheckResourceAttr(res1.TFID(), "ttl", "3600"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.#", "2"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.0.address", "201.34.67.42"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.1.address", "201.34.67.43"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.0.port", "53"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.1.port", "5353"),
					resource.TestCheckResourceAttr(res1.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(res1.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(res1.TFID(), "registrar", "other"),
				),
			},
			{
				ResourceName:      res1.TFID(),
				ImportStateId:     res1.Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_zone", res2),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res1.TFID(), zone.GetAPIResource()),
					resource.TestCheckResourceAttr(res1.TFID(), "name", res2.Name),
					resource.TestCheckResourceAttr(res1.TFID(), "mode", "secondary"),
					resource.TestCheckResourceAttr(res1.TFID(), "labels.key", "value"),
					resource.TestCheckResourceAttr(res1.TFID(), "ttl", "10800"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.#", "2"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.0.address", "89.34.67.42"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.1.address", "89.34.67.43"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.0.port", "53"),
					resource.TestCheckResourceAttr(res1.TFID(), "primary_nameservers.1.port", "5353"),
					resource.TestCheckResourceAttr(res1.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(res1.TFID(), "authoritative_nameservers.assigned.#", "3"),
					resource.TestCheckResourceAttr(res1.TFID(), "registrar", "other"),
				),
			},
		},
	})
}
