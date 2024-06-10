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
					resource.TestCheckResourceAttr(stByName.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(stByName.TFID(), "included_traffic", "21990232555520"),

					resource.TestCheckResourceAttr(stByID.TFID(), "id", "22"),
					resource.TestCheckResourceAttr(stByID.TFID(), "name", "cpx11"),
					resource.TestCheckResourceAttr(stByID.TFID(), "cores", "2"),
					resource.TestCheckResourceAttr(stByID.TFID(), "memory", "2"),
					resource.TestCheckResourceAttr(stByID.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(stByID.TFID(), "included_traffic", "21990232555520"),
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
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_type_ids.0", "1"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "server_type_ids.1", "3"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "names.0", "cx11"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "names.1", "cx21"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "descriptions.0", "CX11"),
					resource.TestCheckResourceAttr(servertypesD.TFID(), "descriptions.1", "CX21"),
				),
			},
		},
	})
}
