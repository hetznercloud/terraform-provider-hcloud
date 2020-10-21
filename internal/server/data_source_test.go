package server_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceServerTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &server.RData{
		Name:  "server-ds-test",
		Type:  "cx11",
		Image: "ubuntu-20.04",
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
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
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
