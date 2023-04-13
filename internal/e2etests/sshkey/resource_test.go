package sshkey_test

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestSSHKeyResource_Basic(t *testing.T) {
	var sk hcloud.SSHKey

	tmplMan := testtemplate.Manager{}
	res := sshkey.NewRData(t, "basic-ssh-key")
	resRenamed := &sshkey.RData{Name: res.Name + "-renamed", PublicKey: res.PublicKey}
	resRenamed.SetRName(res.Name)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
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
