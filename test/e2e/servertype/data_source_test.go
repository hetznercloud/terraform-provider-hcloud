package servertype

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestServerTypeDataSource(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		tmplMan := testtemplate.Manager{}

		byName := &DData{ServerTypeName: teste2e.TestServerType}
		byName.SetRName("by_name")

		byID := &DData{ServerTypeID: "22"}
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
						resource.TestCheckResourceAttr(byName.TFID(), "disk", "40"),
						resource.TestCheckResourceAttr(byName.TFID(), "storage_type", "local"),
						resource.TestCheckResourceAttr(byName.TFID(), "cpu_type", "shared"),
						resource.TestCheckResourceAttr(byName.TFID(), "architecture", "x86"),
						resource.TestCheckResourceAttr(byName.TFID(), "included_traffic", "0"),
						resource.TestCheckResourceAttr(byName.TFID(), "is_deprecated", "false"),
						resource.TestCheckResourceAttr(byName.TFID(), "deprecation_announced", ""),
						resource.TestCheckResourceAttr(byName.TFID(), "unavailable_after", ""),

						resource.TestCheckResourceAttr(byID.TFID(), "id", "22"),
						resource.TestCheckResourceAttr(byID.TFID(), "name", "cpx11"),
						resource.TestCheckResourceAttr(byID.TFID(), "cores", "2"),
						resource.TestCheckResourceAttr(byID.TFID(), "memory", "2"),
						resource.TestCheckResourceAttr(byID.TFID(), "disk", "40"),
						resource.TestCheckResourceAttr(byID.TFID(), "storage_type", "local"),
						resource.TestCheckResourceAttr(byID.TFID(), "cpu_type", "shared"),
						resource.TestCheckResourceAttr(byID.TFID(), "architecture", "x86"),
						resource.TestCheckResourceAttr(byID.TFID(), "included_traffic", "0"),
						resource.TestCheckResourceAttr(byID.TFID(), "is_deprecated", "false"),
						resource.TestCheckResourceAttr(byID.TFID(), "deprecation_announced", ""),
						resource.TestCheckResourceAttr(byID.TFID(), "unavailable_after", ""),
					),
				},
			},
		})

		t.Run("upgrade-plugin-framework", func(t *testing.T) {
			tmplMan := testtemplate.Manager{}

			byName := &DData{ServerTypeName: teste2e.TestServerType}
			byName.SetRName("by_name")

			byID := &DData{ServerTypeID: "22"}
			byID.SetRName("by_id")

			resource.ParallelTest(t, resource.TestCase{
				PreCheck: teste2e.PreCheck(t),
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"hcloud": {
								VersionConstraint: "1.48.1",
								Source:            "hetznercloud/hcloud",
							},
						},

						Config: tmplMan.Render(t,
							"testdata/d/hcloud_server_type", byName,
							"testdata/d/hcloud_server_type", byID,
							"testdata/r/terraform_data_resource", byName,
							"testdata/r/terraform_data_resource", byID,
						),
					},
					{
						ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

						Config: tmplMan.Render(t,
							"testdata/d/hcloud_server_type", byName,
							"testdata/d/hcloud_server_type", byID,
							"testdata/r/terraform_data_resource", byName,
							"testdata/r/terraform_data_resource", byID,
						),

						PlanOnly: true,
					},
				},
			})
		})
	})

	t.Run("list", func(t *testing.T) {
		tmplMan := testtemplate.Manager{}

		all := &DDataList{}
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

		t.Run("upgrade-plugin-framework", func(t *testing.T) {
			tmplMan := testtemplate.Manager{}

			all := &DDataList{}
			all.SetRName("all")

			resource.ParallelTest(t, resource.TestCase{
				PreCheck: teste2e.PreCheck(t),
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"hcloud": {
								VersionConstraint: "1.48.1",
								Source:            "hetznercloud/hcloud",
							},
						},

						Config: tmplMan.Render(t,
							"testdata/d/hcloud_server_types", all,
							"testdata/r/terraform_data_resource", all,
						),
					},
					{
						ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

						Config: tmplMan.Render(t,
							"testdata/d/hcloud_server_types", all,
							"testdata/r/terraform_data_resource", all,
						),

						PlanOnly: true,
					},
				},
			})
		})
	})
}
