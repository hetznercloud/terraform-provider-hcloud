package sshkey_test

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/sshkey"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceSSHKeyTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := sshkey.NewRData(t, "datasource-test")
	sshKeyByName := &sshkey.DData{
		SSHKeyName: res.TFID() + ".name",
	}
	sshKeyByName.SetRName("sshkey_by_name")
	sshKeyByID := &sshkey.DData{
		SSHKeyID: res.TFID() + ".id",
	}
	sshKeyByID.SetRName("sshkey_by_id")
	sshKeyBySel := &sshkey.DData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	sshKeyBySel.SetRName("sshkey_by_sel")

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
					"testdata/d/hcloud_ssh_key", sshKeyByName,
					"testdata/d/hcloud_ssh_key", sshKeyByID,
					"testdata/d/hcloud_ssh_key", sshKeyBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sshKeyByName.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(sshKeyByName.TFID(), "public_key", res.PublicKey),

					resource.TestCheckResourceAttr(sshKeyByID.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(sshKeyByID.TFID(), "public_key", res.PublicKey),

					resource.TestCheckResourceAttr(sshKeyBySel.TFID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(sshKeyBySel.TFID(), "public_key", res.PublicKey),
				),
			},
		},
	})
}
