package sshkey_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
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

func TestAccHcloudDataSourceSSHKeysTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := sshkey.NewRData(t, "datasource-test")

	sshKeysDS := &sshkey.SSHKeysDData{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	sshKeysDS.SetRName("ds")
	resource.Test(t, resource.TestCase{
		PreCheck:  testsupport.AccTestPreCheck(t),
		Providers: testsupport.AccTestProviders(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
					"testdata/d/hcloud_ssh_keys", sshKeysDS,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sshKeysDS.TFID(), "ssh_keys.0.name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(sshKeysDS.TFID(), "ssh_keys.0.public_key", res.PublicKey),
				),
			},
		},
	})
}
