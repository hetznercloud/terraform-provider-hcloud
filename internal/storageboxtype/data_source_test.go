package storageboxtype_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/storageboxtype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccStorageBoxTypeDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	byID := &storageboxtype.DData{
		ID: "1333",
	}
	byID.SetRName("by_id")
	byName := &storageboxtype.DData{
		Name: "bx11",
	}
	byName.SetRName("by_name")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_storage_box_type", byID,
					"testdata/d/hcloud_storage_box_type", byName,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(byID.TFID(), "id", "1333"),
					resource.TestCheckResourceAttr(byID.TFID(), "name", "bx11"),
					resource.TestCheckResourceAttr(byID.TFID(), "description", "BX11"),
					resource.TestCheckResourceAttr(byID.TFID(), "snapshot_limit", "10"),
					resource.TestCheckResourceAttr(byID.TFID(), "automatic_snapshot_limit", "10"),
					resource.TestCheckResourceAttr(byID.TFID(), "subaccounts_limit", "100"),
					resource.TestCheckResourceAttr(byID.TFID(), "size", "1099511627776"), // 1 TiB
					resource.TestCheckResourceAttr(byID.TFID(), "is_deprecated", "false"),
					resource.TestCheckResourceAttr(byID.TFID(), "deprecation_announced", ""),
					resource.TestCheckResourceAttr(byID.TFID(), "unavailable_after", ""),

					resource.TestCheckResourceAttr(byName.TFID(), "id", "1333"),
					resource.TestCheckResourceAttr(byName.TFID(), "name", "bx11"),
					resource.TestCheckResourceAttr(byName.TFID(), "description", "BX11"),
					resource.TestCheckResourceAttr(byName.TFID(), "snapshot_limit", "10"),
					resource.TestCheckResourceAttr(byName.TFID(), "automatic_snapshot_limit", "10"),
					resource.TestCheckResourceAttr(byName.TFID(), "subaccounts_limit", "100"),
					resource.TestCheckResourceAttr(byName.TFID(), "size", "1099511627776"), // 1 TiB
					resource.TestCheckResourceAttr(byName.TFID(), "is_deprecated", "false"),
					resource.TestCheckResourceAttr(byName.TFID(), "deprecation_announced", ""),
					resource.TestCheckResourceAttr(byName.TFID(), "unavailable_after", ""),
				),
			},
		},
	})
}
