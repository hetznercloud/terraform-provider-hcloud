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
					resource.TestCheckResourceAttr(byName.TFID(), "cores", "2"),
					resource.TestCheckResourceAttr(byName.TFID(), "memory", "2"),
					resource.TestCheckResourceAttr(byName.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(byName.TFID(), "included_traffic", "21990232555520"),

					resource.TestCheckResourceAttr(byID.TFID(), "id", "22"),
					resource.TestCheckResourceAttr(byID.TFID(), "name", "cpx11"),
					resource.TestCheckResourceAttr(byID.TFID(), "cores", "2"),
					resource.TestCheckResourceAttr(byID.TFID(), "memory", "2"),
					resource.TestCheckResourceAttr(byID.TFID(), "architecture", "x86"),
					resource.TestCheckResourceAttr(byID.TFID(), "included_traffic", "21990232555520"),
				),
			},
		},
	})
}

func TestAccHcloudDataSourceServerTypesTest(t *testing.T) {
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
					resource.TestCheckResourceAttr(all.TFID(), "server_type_ids.0", "1"),
					resource.TestCheckResourceAttr(all.TFID(), "server_type_ids.1", "3"),
					resource.TestCheckResourceAttr(all.TFID(), "names.0", "cx11"),
					resource.TestCheckResourceAttr(all.TFID(), "names.1", "cx21"),
					resource.TestCheckResourceAttr(all.TFID(), "descriptions.0", "CX11"),
					resource.TestCheckResourceAttr(all.TFID(), "descriptions.1", "CX21"),
				),
			},
		},
	})
}
