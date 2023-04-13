package primaryip_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfhcloud "github.com/hetznercloud/terraform-provider-hcloud/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourcePrimaryIPTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &primaryip.RData{
		Name: "primaryip-ds-test",
		Type: "ipv4",
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
		Datacenter:   "fsn1-dc14",
		AssigneeType: "server",
	}
	res.SetRName("primaryip-ds-test")
	primaryIPByName := &primaryip.DData{
		PrimaryIPName: res.TFID() + ".name",
	}
	primaryIPByName.SetRName("primaryip_by_name")
	primaryIPByID := &primaryip.DData{
		PrimaryIPID: res.TFID() + ".id",
	}

	primaryIPByIP := &primaryip.DData{
		PrimaryIPIP: res.TFID() + ".ip_address",
	}
	primaryIPByIP.SetRName("primaryip_by_ip")

	primaryIPByID.SetRName("primaryip_by_id")
	primaryIPBySel := &primaryip.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	primaryIPBySel.SetRName("primaryip_by_sel")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: e2etests.PreCheck(t),
		ProviderFactories: map[string]func() (*schema.Provider, error){
			//nolint:unparam
			"hcloud": func() (*schema.Provider, error) {
				return tfhcloud.Provider(), nil
			},
		},
		CheckDestroy: testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
					"testdata/d/hcloud_primary_ip", primaryIPByName,
					"testdata/d/hcloud_primary_ip", primaryIPByIP,
					"testdata/d/hcloud_primary_ip", primaryIPByID,
					"testdata/d/hcloud_primary_ip", primaryIPBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(primaryIPByName.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(primaryIPByName.TFID(), "datacenter", res.Datacenter),

					resource.TestCheckResourceAttr(primaryIPByID.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(primaryIPByID.TFID(), "datacenter", res.Datacenter),

					resource.TestCheckResourceAttr(primaryIPBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(primaryIPBySel.TFID(), "datacenter", res.Datacenter),
				),
			},
		},
	})
}

func TestAccHcloudDataSourcePrimaryIPListTest(t *testing.T) {
	res := &primaryip.RData{
		Name: "primary-ds-test",
		Type: "ipv4",
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
		Datacenter:   "fsn1-dc14",
		AssigneeType: "server",
	}
	res.SetRName("primary-ds-test")

	primaryipBySel := &primaryip.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	primaryipBySel.SetRName("primaryip_by_sel")

	allPrimaryIPsSel := &primaryip.DDataList{}
	allPrimaryIPsSel.SetRName("all_primaryips_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: e2etests.PreCheck(t),
		ProviderFactories: map[string]func() (*schema.Provider, error){
			//nolint:unparam
			"hcloud": func() (*schema.Provider, error) {
				return tfhcloud.Provider(), nil
			},
		},
		CheckDestroy: testsupport.CheckResourcesDestroyed(primaryip.ResourceType, primaryip.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_primary_ip", res,
					"testdata/d/hcloud_primary_ips", primaryipBySel,
					"testdata/d/hcloud_primary_ips", allPrimaryIPsSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(primaryipBySel.TFID(), "primary_ips.*",
						map[string]string{
							"name":       fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"datacenter": res.Datacenter,
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(allPrimaryIPsSel.TFID(), "primary_ips.*",
						map[string]string{
							"name":       fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"datacenter": res.Datacenter,
						},
					),
				),
			},
		},
	})
}
