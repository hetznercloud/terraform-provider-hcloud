package primaryip_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/kit/randutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccPrimaryIPDataSourceList(t *testing.T) {
	res := &primaryip.RData{
		Name:         "main",
		Type:         "ipv6",
		Location:     teste2e.TestLocationName,
		AssigneeType: "server",
		Labels:       map[string]string{"key": randutil.GenerateID()},
	}
	res.SetRName("main")

	byLabel := &primaryip.DDataList{
		LabelSelector: fmt.Sprintf("key=%s", res.Labels["key"]),
		Raw:           fmt.Sprintf("depends_on = [%s]", res.TFID()),
	}
	byLabel.SetRName("by_label")

	all := &primaryip.DDataList{}
	all.SetRName("all")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
					"testdata/d/hcloud_primary_ips", byLabel,
					"testdata/d/hcloud_primary_ips", all,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("primary_ips").AtSliceIndex(0).AtMapKey("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("primary_ips").AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("primary_ips").AtSliceIndex(0).AtMapKey("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("primary_ips").AtSliceIndex(0).AtMapKey("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("primary_ips").AtSliceIndex(0).AtMapKey("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("primary_ips").AtSliceIndex(0).AtMapKey("assignee_type"), knownvalue.StringExact("server")),

					statecheck.ExpectKnownValue(all.TFID(), tfjsonpath.New("primary_ips"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccPrimaryIPDataSourceList_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &primaryip.RData{
		Name:         "main",
		Type:         "ipv6",
		Location:     teste2e.TestLocationName,
		AssigneeType: "server",
		Labels:       map[string]string{"key": randutil.GenerateID()},
	}
	res.SetRName("main")

	byLabel := &primaryip.DDataList{
		LabelSelector: fmt.Sprintf("key=%s", res.Labels["key"]),
		Raw:           fmt.Sprintf("depends_on = [%s]", res.TFID()),
	}
	byLabel.SetRName("by_label")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.60.1",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
				),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.60.1",
						Source:            "hetznercloud/hcloud",
					},
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
					"testdata/d/hcloud_primary_ips", byLabel,
					"testdata/r/terraform_data_resource", byLabel,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
					"testdata/d/hcloud_primary_ips", byLabel,
					"testdata/r/terraform_data_resource", byLabel,
				),
				PlanOnly: true,
			},
		},
	})
}
