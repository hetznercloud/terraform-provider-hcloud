package sshkey_test

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/sshkey"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestSSHKeyResource_Basic(t *testing.T) {
	var sk hcloud.SSHKey

	res := sshkey.NewRData(t, "basic-ssh-key")
	resRenamed := &sshkey.RData{Name: res.Name + "-renamed", PublicKey: res.PublicKey}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(sshkey.ResourceType, sshkey.ByID(t, &sk)),
		Steps: []resource.TestStep{
			{
				// Create a new SSH Key using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_ssh_key", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), sshkey.ByID(t, &sk)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("basic-ssh-key--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "public_key", res.PublicKey),
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
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("basic-ssh-key-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "public_key", res.PublicKey),
				),
			},
		},
	})
}
