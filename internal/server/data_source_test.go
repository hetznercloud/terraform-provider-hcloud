package server_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccServerDataSource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &server.RData{
		Name:  "server-ds-test",
		Type:  teste2e.TestServerType,
		Image: teste2e.TestImage,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	res.SetRName("server-ds-test")
	serverByName := &server.DData{
		ServerName: res.TFID() + ".name",
	}
	serverByName.SetRName("server_by_name")
	serverByID := &server.DData{
		ServerID: res.TFID() + ".id",
	}
	serverByID.SetRName("server_by_id")

	serverBySel := &server.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	serverBySel.SetRName("server_by_sel")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res,
					"testdata/d/hcloud_server", serverByName,
					"testdata/d/hcloud_server", serverByID,
					"testdata/d/hcloud_server", serverBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(serverByName.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),

					resource.TestCheckResourceAttr(serverByID.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(serverBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
				),
			},
		},
	})
}

func TestAccServerDataSourceList(t *testing.T) {
	res := &server.RData{
		Name:  "server-ds-test",
		Type:  teste2e.TestServerType,
		Image: teste2e.TestImage,
		Labels: map[string]string{
			"key": strconv.Itoa(acctest.RandInt()),
		},
	}
	res.SetRName("server-ds-test")

	serversBySel := &server.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	serversBySel.SetRName("server_by_sel")

	allServersSel := &server.DDataList{}
	allServersSel.SetRName("all_servers_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res,
					"testdata/d/hcloud_servers", serversBySel,
					"testdata/d/hcloud_servers", allServersSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(serversBySel.TFID(), "servers.*",
						map[string]string{
							"name": fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(allServersSel.TFID(), "servers.*",
						map[string]string{
							"name": fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
						},
					),
				),
			},
		},
	})
}
