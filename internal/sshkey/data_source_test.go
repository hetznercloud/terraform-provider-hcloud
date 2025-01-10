package sshkey_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccSSHKeyDataSource(t *testing.T) {
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
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(sshkey.ResourceType, sshkey.ByID(t, nil)),
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

func TestAccSSHKeyDataSource_UpgradePluginFramework(t *testing.T) {
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
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.44.1",
						Source:            "hetznercloud/hcloud",
					},
				},

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
				),
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.44.1",
						Source:            "hetznercloud/hcloud",
					},
				},

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
					"testdata/d/hcloud_ssh_key", sshKeyByName,
					"testdata/d/hcloud_ssh_key", sshKeyByID,
					"testdata/d/hcloud_ssh_key", sshKeyBySel,

					"testdata/r/terraform_data_resource", sshKeyByName,
					"testdata/r/terraform_data_resource", sshKeyByID,
					"testdata/r/terraform_data_resource", sshKeyBySel,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
					"testdata/d/hcloud_ssh_key", sshKeyByName,
					"testdata/d/hcloud_ssh_key", sshKeyByID,
					"testdata/d/hcloud_ssh_key", sshKeyBySel,

					"testdata/r/terraform_data_resource", sshKeyByName,
					"testdata/r/terraform_data_resource", sshKeyByID,
					"testdata/r/terraform_data_resource", sshKeyBySel,
				),

				PlanOnly: true,
			},
		},
	})
}
