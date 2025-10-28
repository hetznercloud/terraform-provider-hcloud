package storageboxtype_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxtype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxTypeDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &storageboxtype.DDataList{}
	res.SetRName("all")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_storage_box_types", res,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.#", "4"),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.name", "bx11"),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.description", "BX11"),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.snapshot_limit", "10"),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.automatic_snapshot_limit", "10"),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.subaccounts_limit", "100"),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.size", "1099511627776"), // 1 TiB
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.is_deprecated", "false"),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.deprecation_announced", ""),
					resource.TestCheckResourceAttr(res.TFID(), "storage_box_types.0.unavailable_after", ""),
				),
			},
		},
	})
}
