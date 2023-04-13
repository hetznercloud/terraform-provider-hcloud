package sshkey_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(sshkey.ResourceType, sshkey.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
				),
			},
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
	res := sshkey.NewRData(t, "ssh-key-ds-test")

	keyBySel := &sshkey.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	keyBySel.SetRName("key_by_sel")

	allKeysSel := &sshkey.DDataList{}
	allKeysSel.SetRName("all_keys_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  e2etests.PreCheck(t),
		Providers: e2etests.Providers(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
					"testdata/d/hcloud_ssh_keys", keyBySel,
					"testdata/d/hcloud_ssh_keys", allKeysSel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(keyBySel.TFID(), "ssh_keys.*",
						map[string]string{
							"name":       fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"public_key": res.PublicKey,
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(allKeysSel.TFID(), "ssh_keys.*",
						map[string]string{
							"name":       fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"public_key": res.PublicKey,
						},
					),
				),
			},
		},
	})
}
