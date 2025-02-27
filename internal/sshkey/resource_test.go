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

func TestAccSSHKeyResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := sshkey.NewRData(t, "basic-ssh-key")
	resRenamed := &sshkey.RData{Name: res.Name + "-renamed", PublicKey: res.PublicKey}
	resRenamed.SetRName(res.Name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(sshkey.ResourceType, sshkey.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new SSH Key using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_ssh_key", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res.TFID(), sshkey.GetAPIResource()),
					resource.TestCheckResourceAttr(res.TFID(), "name", fmt.Sprintf("basic-ssh-key--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "public_key", res.PublicKey),
					resource.TestCheckResourceAttrSet(res.TFID(), "fingerprint"),
					resource.TestCheckResourceAttr(res.TFID(), "labels.key", res.Labels["key"]),
				),
			},
			{
				// Try to import the newly created SSH Key
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the SSH Key created in the previous step by
				// setting all optional fields and renaming the SSH
				// Key.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name", fmt.Sprintf("basic-ssh-key-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "public_key", res.PublicKey),
				),
			},
		},
	})
}

func TestAccSSHKeyResource_UpgradePluginFramework(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := sshkey.NewRData(t, "upgrade-plugin-framework-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: teste2e.PreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hcloud": {
						VersionConstraint: "1.45.0",
						Source:            "hetznercloud/hcloud",
					},
				},

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
				),
			},
			{
				ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", res,
				),

				PlanOnly: true,
			},
		},
	})
}
