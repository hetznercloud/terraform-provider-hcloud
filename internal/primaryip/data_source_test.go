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

func TestAccPrimaryIPDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &primaryip.RData{
		Name:     "main",
		Type:     "ipv6",
		Location: teste2e.TestLocationName,
		Labels:   map[string]string{"key": randutil.GenerateID()},
	}
	res.SetRName("main")

	byName := &primaryip.DData{
		PrimaryIPName: res.TFID() + ".name",
	}
	byName.SetRName("by_name")

	byID := &primaryip.DData{
		PrimaryIPID: res.TFID() + ".id",
	}
	byID.SetRName("by_id")

	byIPAddress := &primaryip.DData{
		PrimaryIPIP: res.TFID() + ".ip_address",
	}
	byIPAddress.SetRName("by_ip_address")

	byLabel := &primaryip.DData{
		LabelSelector: fmt.Sprintf("key=%s", res.Labels["key"]),
		Raw:           fmt.Sprintf("depends_on = [%s]", res.TFID()),
	}
	byLabel.SetRName("by_label")

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
					"testdata/d/hcloud_primary_ip", byID,
					"testdata/d/hcloud_primary_ip", byName,
					"testdata/d/hcloud_primary_ip", byIPAddress,
					"testdata/d/hcloud_primary_ip", byLabel,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(byID.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),

					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(byName.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),

					statecheck.ExpectKnownValue(byIPAddress.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byIPAddress.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byIPAddress.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byIPAddress.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(byIPAddress.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(byIPAddress.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),

					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("name"), knownvalue.StringExact(fmt.Sprintf("main--%d", tmplMan.RandInt))),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("type"), knownvalue.StringExact("ipv6")),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("assignee_id"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(byLabel.TFID(), tfjsonpath.New("assignee_type"), knownvalue.StringExact("server")),
				},
			},
		},
	})
}

func TestAccPrimaryIPDataSource_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &primaryip.RData{
		Name:         "main",
		Type:         "ipv6",
		Location:     teste2e.TestLocationName,
		Labels:       map[string]string{"key": randutil.GenerateID()},
		AssigneeType: "server", // Attribute was still required in previous versions
	}
	res.SetRName("main")

	byName := &primaryip.DData{
		PrimaryIPName: res.TFID() + ".name",
	}
	byName.SetRName("by_name")

	byID := &primaryip.DData{
		PrimaryIPID: res.TFID() + ".id",
	}
	byID.SetRName("by_id")

	byIPAddress := &primaryip.DData{
		PrimaryIPIP: res.TFID() + ".ip_address",
	}
	byIPAddress.SetRName("by_ip_address")

	byLabel := &primaryip.DData{
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
					"testdata/d/hcloud_primary_ip", byID,
					"testdata/d/hcloud_primary_ip", byName,
					"testdata/d/hcloud_primary_ip", byIPAddress,
					"testdata/d/hcloud_primary_ip", byLabel,
					"testdata/r/terraform_data_resource", byID,
					"testdata/r/terraform_data_resource", byName,
					"testdata/r/terraform_data_resource", byIPAddress,
					"testdata/r/terraform_data_resource", byLabel,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
					"testdata/d/hcloud_primary_ip", byID,
					"testdata/d/hcloud_primary_ip", byName,
					"testdata/d/hcloud_primary_ip", byIPAddress,
					"testdata/d/hcloud_primary_ip", byLabel,
					"testdata/r/terraform_data_resource", byID,
					"testdata/r/terraform_data_resource", byName,
					"testdata/r/terraform_data_resource", byIPAddress,
					"testdata/r/terraform_data_resource", byLabel,
				),
				PlanOnly: true,
			},
		},
	})
}
