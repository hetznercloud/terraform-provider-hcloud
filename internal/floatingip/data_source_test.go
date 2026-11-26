package floatingip_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/floatingip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccFloatingIPDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res1 := &floatingip.RData{
		Name:             "main-ipv6",
		Type:             "ipv6",
		HomeLocationName: teste2e.TestLocationName,
		Labels:           map[string]string{"key": "ipv6-" + randutil.GenerateID()},
	}
	res1.SetRName("main-ipv6")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.Name = "main-ipv4"
	res2.Type = "ipv4"
	res2.Labels = map[string]string{"key": "ipv4-" + randutil.GenerateID()}
	res2.SetRName("main-ipv4")

	byName1 := &floatingip.DData{
		FloatingIPName: res1.TFID() + ".name",
	}
	byName1.SetRName("by_name_ipv6")

	byName2 := &floatingip.DData{
		FloatingIPName: res2.TFID() + ".name",
	}
	byName2.SetRName("by_name_ipv4")

	byID1 := &floatingip.DData{
		FloatingIPID: res1.TFID() + ".id",
	}
	byID1.SetRName("by_id_ipv6")

	byID2 := &floatingip.DData{
		FloatingIPID: res2.TFID() + ".id",
	}
	byID2.SetRName("by_id_ipv4")

	byIPAddr1 := &floatingip.DData{
		FloatingIPIP: res1.TFID() + ".ip_address",
	}
	byIPAddr1.SetRName("by_ip_address_ipv6")

	byIPAddr2 := &floatingip.DData{
		FloatingIPIP: res2.TFID() + ".ip_address",
	}
	byIPAddr2.SetRName("by_ip_address_ipv4")

	byLabel1 := &floatingip.DData{
		LabelSelector: fmt.Sprintf("key=%s", res1.Labels["key"]),
		Raw:           fmt.Sprintf("depends_on = [%s]", res1.TFID()),
	}
	byLabel1.SetRName("by_label_ipv6")

	byLabel2 := &floatingip.DData{
		LabelSelector: fmt.Sprintf("key=%s", res2.Labels["key"]),
		Raw:           fmt.Sprintf("depends_on = [%s]", res2.TFID()),
	}
	byLabel2.SetRName("by_label_ipv4")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(floatingip.ResourceType, floatingip.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", res1,
					"testdata/r/hcloud_floating_ip", res2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", res1,
					"testdata/r/hcloud_floating_ip", res2,
					"testdata/d/hcloud_floating_ip", byID1,
					"testdata/d/hcloud_floating_ip", byID2,
					"testdata/d/hcloud_floating_ip", byName1,
					"testdata/d/hcloud_floating_ip", byName2,
					"testdata/d/hcloud_floating_ip", byIPAddr1,
					"testdata/d/hcloud_floating_ip", byIPAddr2,
					"testdata/d/hcloud_floating_ip", byLabel1,
					"testdata/d/hcloud_floating_ip", byLabel2,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byID1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byID1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byID1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byID1.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byID1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byID1.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`::$`))),
					statecheck.ExpectKnownValue(byID1.TFID(), tfjsonpath.New("ip_network"), knownvalue.StringRegexp(regexp.MustCompile(`\/64$`))),

					statecheck.ExpectKnownValue(byID2.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byID2.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					statecheck.ExpectKnownValue(byID2.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byID2.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byID2.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byID2.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`\.\d+$`))),
					statecheck.ExpectKnownValue(byID2.TFID(), tfjsonpath.New("ip_network"), knownvalue.Null()),

					statecheck.ExpectKnownValue(byName1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byName1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byName1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byName1.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byName1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byName1.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`::$`))),
					statecheck.ExpectKnownValue(byName1.TFID(), tfjsonpath.New("ip_network"), knownvalue.StringRegexp(regexp.MustCompile(`\/64$`))),

					statecheck.ExpectKnownValue(byName2.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byName2.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					statecheck.ExpectKnownValue(byName2.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byName2.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byName2.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byName2.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`\.\d+$`))),
					statecheck.ExpectKnownValue(byName2.TFID(), tfjsonpath.New("ip_network"), knownvalue.Null()),

					statecheck.ExpectKnownValue(byIPAddr1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byIPAddr1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byIPAddr1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byIPAddr1.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byIPAddr1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byIPAddr1.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`::$`))),
					statecheck.ExpectKnownValue(byIPAddr1.TFID(), tfjsonpath.New("ip_network"), knownvalue.StringRegexp(regexp.MustCompile(`\/64$`))),

					statecheck.ExpectKnownValue(byIPAddr2.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byIPAddr2.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					statecheck.ExpectKnownValue(byIPAddr2.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byIPAddr2.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byIPAddr2.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byIPAddr2.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`\.\d+$`))),
					statecheck.ExpectKnownValue(byIPAddr2.TFID(), tfjsonpath.New("ip_network"), knownvalue.Null()),

					statecheck.ExpectKnownValue(byLabel1.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv6--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byLabel1.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byLabel1.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byLabel1.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byLabel1.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byLabel1.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`::$`))),
					statecheck.ExpectKnownValue(byLabel1.TFID(), tfjsonpath.New("ip_network"), knownvalue.StringRegexp(regexp.MustCompile(`\/64$`))),

					statecheck.ExpectKnownValue(byLabel2.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main-ipv4--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byLabel2.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv4")),
					statecheck.ExpectKnownValue(byLabel2.TFID(), tfjsonpath.New("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byLabel2.TFID(), tfjsonpath.New("server_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(byLabel2.TFID(), tfjsonpath.New("delete_protection"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(byLabel2.TFID(), tfjsonpath.New("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`\.\d+$`))),
					statecheck.ExpectKnownValue(byLabel2.TFID(), tfjsonpath.New("ip_network"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccFloatingIPDataSourceList(t *testing.T) {
	res := &floatingip.RData{
		Name:             "main",
		Type:             "ipv6",
		HomeLocationName: teste2e.TestLocationName,
		Labels:           map[string]string{"key": randutil.GenerateID()},
	}
	res.SetRName("main")

	byLabel := &floatingip.DDataList{
		LabelSelector: fmt.Sprintf("key=%s", res.Labels["key"]),
		Raw:           fmt.Sprintf("depends_on = [%s]", res.TFID()),
	}
	byLabel.SetRName("by_label_ipv6")

	all := &floatingip.DDataList{}
	all.SetRName("all")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(floatingip.ResourceType, floatingip.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_floating_ip", res,
					"testdata/d/hcloud_floating_ips", byLabel,
					"testdata/d/hcloud_floating_ips", all,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("floating_ips").AtSliceIndex(0).AtMapKey("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("floating_ips").AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("floating_ips").AtSliceIndex(0).AtMapKey("home_location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("floating_ips").AtSliceIndex(0).AtMapKey("server_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("floating_ips").AtSliceIndex(0).AtMapKey("ip_address"), knownvalue.StringRegexp(regexp.MustCompile(`::$`))),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("floating_ips").AtSliceIndex(0).AtMapKey("ip_network"), knownvalue.StringRegexp(regexp.MustCompile(`\/64$`))),

					statecheck.ExpectKnownValue(all.TFID(), tfjsonpath.New("floating_ips"), knownvalue.NotNull()),
				},
			},
		},
	})
}
