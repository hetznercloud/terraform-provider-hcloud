package sshkey_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccSSHKeyDataSourceList(t *testing.T) {
	res := sshkey.NewRData(t, "ssh-key-ds-test")

	sshKeysByLabelSelector := &sshkey.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	sshKeysByLabelSelector.SetRName("key_by_sel")

	sshKeysAll := &sshkey.DDataList{}
	sshKeysAll.SetRName("all_keys_sel")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
					"testdata/d/hcloud_ssh_keys", sshKeysByLabelSelector,
					"testdata/d/hcloud_ssh_keys", sshKeysAll,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(sshKeysByLabelSelector.TFID(), "ssh_keys.*",
						map[string]string{
							"name":       fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt),
							"public_key": res.PublicKey,
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(sshKeysAll.TFID(), "ssh_keys.*",
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

func TestAccSSHKeyDataSourceList_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := sshkey.NewRData(t, "ssh-key-ds-test")

	sshKeysByLabelSelector := &sshkey.DDataList{
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.TFID()),
	}
	sshKeysByLabelSelector.SetRName("key_by_sel")

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

					"testdata/d/hcloud_ssh_keys", sshKeysByLabelSelector,

					"testdata/r/terraform_data_resource", sshKeysByLabelSelector,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,

					"testdata/d/hcloud_ssh_keys", sshKeysByLabelSelector,

					"testdata/r/terraform_data_resource", sshKeysByLabelSelector,
				),

				PlanOnly: true,
			},
		},
	})
}
