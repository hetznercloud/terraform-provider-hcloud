package servertype_test

import (
	"context"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/servertype"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccDataSource(t *testing.T) {
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
}

func TestAccDataSourceList(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	all := &servertype.DDataList{}
	all.SetRName("all")

	// To make this test more resilient against changes to the server types list, we
	// pre-calculate the indices of the server types we check against.

	// Acceptance tests are usually skipped by terraform, we need to handle this ourselves here.
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", resource.EnvTfAcc)
		return
	}
	// Make sure that a token is available
	teste2e.PreCheck(t)
	client, err := testsupport.CreateClient()
	if err != nil {
		t.Fatal(err.Error())
	}
	serverTypes, err := client.ServerType.All(context.Background())
	if err != nil {
		t.Fatalf("failed to get list of server types: %v", err)
	}

	indexCPX11 := slices.IndexFunc(serverTypes, func(serverType *hcloud.ServerType) bool { return serverType.Name == "cpx11" })
	indexCPX21 := slices.IndexFunc(serverTypes, func(serverType *hcloud.ServerType) bool { return serverType.Name == "cpx21" })

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/d/hcloud_server_types", all,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_type_ids.%d", indexCPX11), "22"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_type_ids.%d", indexCPX21), "23"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("names.%d", indexCPX11), "cpx11"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("names.%d", indexCPX21), "cpx21"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("descriptions.%d", indexCPX11), "CPX 11"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("descriptions.%d", indexCPX21), "CPX 21"),

					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.id", indexCPX11), "22"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.name", indexCPX11), "cpx11"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.cores", indexCPX11), "2"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.memory", indexCPX11), "2"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.disk", indexCPX11), "40"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.storage_type", indexCPX11), "local"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.cpu_type", indexCPX11), "shared"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.architecture", indexCPX11), "x86"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.included_traffic", indexCPX11), "0"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.is_deprecated", indexCPX11), "false"),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.deprecation_announced", indexCPX11), ""),
					resource.TestCheckResourceAttr(all.TFID(), fmt.Sprintf("server_types.%d.unavailable_after", indexCPX11), ""),
				),
			},
		},
	})
}
