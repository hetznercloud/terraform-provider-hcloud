package servertype_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/servertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccServerTypeDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	byName := &servertype.DData{ServerTypeName: teste2e.TestServerType}
	byName.SetRName("by_name")

	byID := &servertype.DData{ServerTypeID: "22"}
	byID.SetRName("by_id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_server_type", byName,
					"testdata/d/hcloud_server_type", byID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(byName.TFID(), "id", "22"),
					resource.TestCheckResourceAttr(byName.TFID(), "name", "cpx11"),
					resource.TestCheckResourceAttr(byName.TFID(), "category", "Shared vCPU"),
					resource.TestCheckResourceAttr(byName.TFID(), "cores", "2"),
					resource.TestCheckResourceAttr(byName.TFID(), "memory", "2"),
					resource.TestCheckResourceAttr(byName.TFID(), "disk", "40"),
					resource.TestCheckResourceAttr(byName.TFID(), "storage_type", "local"),
					resource.TestCheckResourceAttr(byName.TFID(), "cpu_type", "shared"),
					resource.TestCheckResourceAttr(byName.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.#", "6"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.id", "1"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.name", "fsn1"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.is_deprecated", "false"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.deprecation_announced", ""),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.unavailable_after", ""),

					resource.TestCheckResourceAttr(byName.TFID(), "included_traffic", "0"),
					resource.TestCheckResourceAttr(byName.TFID(), "is_deprecated", "false"),
					resource.TestCheckResourceAttr(byName.TFID(), "deprecation_announced", ""),
					resource.TestCheckResourceAttr(byName.TFID(), "unavailable_after", ""),

					resource.TestCheckResourceAttr(byID.TFID(), "id", "22"),
					resource.TestCheckResourceAttr(byID.TFID(), "name", "cpx11"),
					resource.TestCheckResourceAttr(byID.TFID(), "category", "Shared vCPU"),
					resource.TestCheckResourceAttr(byID.TFID(), "cores", "2"),
					resource.TestCheckResourceAttr(byID.TFID(), "memory", "2"),
					resource.TestCheckResourceAttr(byID.TFID(), "disk", "40"),
					resource.TestCheckResourceAttr(byID.TFID(), "storage_type", "local"),
					resource.TestCheckResourceAttr(byID.TFID(), "cpu_type", "shared"),
					resource.TestCheckResourceAttr(byID.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.#", "6"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.id", "1"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.name", "fsn1"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.is_deprecated", "false"),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.deprecation_announced", ""),
					resource.TestCheckResourceAttr(byName.TFID(), "locations.0.unavailable_after", ""),

					resource.TestCheckResourceAttr(byID.TFID(), "included_traffic", "0"),
					resource.TestCheckResourceAttr(byID.TFID(), "is_deprecated", "false"),
					resource.TestCheckResourceAttr(byID.TFID(), "deprecation_announced", ""),
					resource.TestCheckResourceAttr(byID.TFID(), "unavailable_after", ""),
				),
			},
		},
	})
}

func TestAccServerTypeDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	all := &servertype.DDataList{}
	all.SetRName("all")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_server_types", all,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(all.TFID(), "server_type_ids.0", "22"),
					resource.TestCheckResourceAttr(all.TFID(), "server_type_ids.1", "23"),
					resource.TestCheckResourceAttr(all.TFID(), "names.0", "cpx11"),
					resource.TestCheckResourceAttr(all.TFID(), "names.1", "cpx21"),
					resource.TestCheckResourceAttr(all.TFID(), "descriptions.0", "CPX 11"),
					resource.TestCheckResourceAttr(all.TFID(), "descriptions.1", "CPX 21"),

					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.id", "22"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.name", "cpx11"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.cores", "2"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.memory", "2"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.disk", "40"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.storage_type", "local"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.cpu_type", "shared"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.architecture", "x86"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.included_traffic", "0"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.is_deprecated", "false"),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.deprecation_announced", ""),
					resource.TestCheckResourceAttr(all.TFID(), "server_types.0.unavailable_after", ""),
				),
			},
		},
	})
}
