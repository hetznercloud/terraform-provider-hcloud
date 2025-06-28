package iso_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/iso"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccISODataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	// Define struct for getting an ISO by name
	isoByName := &iso.DData{
		IsoName: teste2e.TestISO,
	}
	isoByName.SetRName("iso_by_name")

	// Define struct for getting an ISO by name
	isoByNamePrefix := &iso.DData{
		IsoNamePrefix: teste2e.TestISO[:10],
	}
	isoByName.SetRName("iso_by_name_prefix")

	// Define struct for getting an ISO by ID
	isoByID := &iso.DData{
		IsoID: teste2e.TestIsoID,
	}
	isoByID.SetRName("iso_by_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_iso", isoByName,
					"testdata/d/hcloud_iso", isoByNamePrefix,
					"testdata/d/hcloud_iso", isoByID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(isoByName.TFID(),
						"name", teste2e.TestISO),
					resource.TestCheckResourceAttr(isoByName.TFID(), "id", teste2e.TestIsoID),

					resource.TestCheckResourceAttr(isoByNamePrefix.TFID(),
						"name", teste2e.TestISO),
					resource.TestCheckResourceAttr(isoByNamePrefix.TFID(), "id", teste2e.TestIsoID),

					resource.TestCheckResourceAttr(isoByID.TFID(),
						"name", teste2e.TestISO),
					resource.TestCheckResourceAttr(isoByID.TFID(), "id", teste2e.TestIsoID),
				),
			},
		},
	})
}

func TestAccISODataSource_WithFilters(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	isoByNamePrefix := &iso.DData{
		IsoNamePrefix:               teste2e.TestImage[:10],
		Architecture:                "arm",
		IncludeArchitectureWildcard: true,
	}
	isoByNamePrefix.SetRName("iso_by_name_prefix")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_iso", isoByNamePrefix,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(isoByNamePrefix.TFID(),
						"name", teste2e.TestISO),
					resource.TestCheckResourceAttr(isoByNamePrefix.TFID(),
						"architecture", "arm"),
				),
			},
		},
	})
}

func TestAccISODataSourceList(t *testing.T) {
	allISOsPrivSel := &iso.DDataList{
		IsoType: hcloud.ISOTypePrivate,
	}
	allISOsPrivSel.SetRName("all_isos_priv_sel")

	allISOsPubSel := &iso.DDataList{
		IsoType: hcloud.ISOTypePublic,
	}
	allISOsPubSel.SetRName("all_isos_pub_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_isos", allISOsPrivSel,
					"testdata/d/hcloud_isos", allISOsPubSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(allISOsPrivSel.TFID(), "images.*",
						map[string]string{
							"name": teste2e.TestISO,
							"id":   teste2e.TestIsoID,
							"type": string(hcloud.ISOTypePrivate),
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(allISOsPubSel.TFID(), "images.*",
						map[string]string{
							"name": teste2e.TestISO,
							"id":   teste2e.TestIsoID,
							"type": string(hcloud.ISOTypePublic),
						},
					),
				),
			},
		},
	})
}
