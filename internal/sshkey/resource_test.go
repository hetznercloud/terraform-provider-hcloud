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

	res := sshkey.NewRData(t, "main")

	res2 := &sshkey.RData{Name: res.Name, PublicKey: res.PublicKey}
	res2.SetRName(res.RName())

	res3 := &sshkey.RData{Name: res.Name + "-renamed", PublicKey: res.PublicKey, Labels: res.Labels}
	res3.SetRName(res.RName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(sshkey.ResourceType, sshkey.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_ssh_key", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res.TFID(), sshkey.GetAPIResource()),
					resource.TestCheckResourceAttr(res.TFID(), "name", fmt.Sprintf("main--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "public_key", res.PublicKey),
					resource.TestCheckResourceAttrSet(res.TFID(), "fingerprint"),
					resource.TestCheckResourceAttr(res.TFID(), "labels.key", res.Labels["key"]),
				),
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_ssh_key", res2),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res2.TFID(), sshkey.GetAPIResource()),
					resource.TestCheckResourceAttr(res2.TFID(), "name", fmt.Sprintf("main--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res2.TFID(), "public_key", res2.PublicKey),
					resource.TestCheckResourceAttrSet(res2.TFID(), "fingerprint"),
					resource.TestCheckResourceAttr(res2.TFID(), "labels.key.#", "0"),
				),
			},
			{
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: tmplMan.Render(t, "testdata/r/hcloud_ssh_key", res3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res3.TFID(), "name", fmt.Sprintf("main-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res3.TFID(), "public_key", res3.PublicKey),
					resource.TestCheckResourceAttrSet(res3.TFID(), "fingerprint"),
					resource.TestCheckResourceAttr(res3.TFID(), "labels.key", res3.Labels["key"]),
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
