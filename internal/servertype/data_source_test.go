package servertype_test

import (
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/servertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceServerTypeTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	stByName := &servertype.DData{
		ServerTypeName: teste2e.TestServerType,
	}
	stByName.SetRName("st_by_name")
	stByID := &servertype.DData{
		ServerTypeID: "22",
	}
	stByID.SetRName("st_by_id")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_server_type", stByName,
					"testdata/d/hcloud_server_type", stByID,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(stByName.TFID(), "id", "22"),
					resource.TestCheckResourceAttr(stByName.TFID(), "name", "cpx11"),
					resource.TestCheckResourceAttr(stByName.TFID(), "cores", "2"),
					resource.TestCheckResourceAttr(stByName.TFID(), "memory", "2"),
					resource.TestCheckResourceAttr(stByName.TFID(), "disk", "40"),
					resource.TestCheckResourceAttr(stByName.TFID(), "storage_type", "local"),
					resource.TestCheckResourceAttr(stByName.TFID(), "cpu_type", "shared"),
					resource.TestCheckResourceAttr(stByName.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(stByName.TFID(), "included_traffic", "0"),
					resource.TestCheckResourceAttr(stByName.TFID(), "is_deprecated", "false"),
					resource.TestCheckResourceAttr(stByName.TFID(), "deprecation_announced", ""),
					resource.TestCheckResourceAttr(stByName.TFID(), "unavailable_after", ""),

					resource.TestCheckResourceAttr(stByID.TFID(), "id", "22"),
					resource.TestCheckResourceAttr(stByID.TFID(), "name", "cpx11"),
					resource.TestCheckResourceAttr(stByID.TFID(), "cores", "2"),
					resource.TestCheckResourceAttr(stByID.TFID(), "memory", "2"),
					resource.TestCheckResourceAttr(stByID.TFID(), "disk", "40"),
					resource.TestCheckResourceAttr(stByID.TFID(), "storage_type", "local"),
					resource.TestCheckResourceAttr(stByID.TFID(), "cpu_type", "shared"),
					resource.TestCheckResourceAttr(stByID.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(stByID.TFID(), "included_traffic", "0"),
					resource.TestCheckResourceAttr(stByID.TFID(), "is_deprecated", "false"),
					resource.TestCheckResourceAttr(stByID.TFID(), "deprecation_announced", ""),
					resource.TestCheckResourceAttr(stByID.TFID(), "unavailable_after", ""),
				),
			},
		},
	})
}

func TestAccHcloudDataSourceServerTypesTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	servertypesD := &servertype.DDataList{}
	servertypesD.SetRName("ds")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_server_types", servertypesD,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_type_ids.0", "22"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_type_ids.1", "23"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "names.0", "cpx11"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "names.1", "cpx21"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "descriptions.0", "CPX 11"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "descriptions.1", "CPX 21"),

					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.id", "22"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.name", "cpx11"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.cores", "2"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.memory", "2"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.disk", "40"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.storage_type", "local"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.cpu_type", "shared"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.architecture", "x86"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.included_traffic", "0"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.is_deprecated", "false"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.deprecation_announced", ""),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_types.0.unavailable_after", ""),
				),
			},
		},
	})
}
