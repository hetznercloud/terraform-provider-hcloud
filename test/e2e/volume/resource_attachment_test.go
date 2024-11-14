package volume

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	servertest "github.com/hetznercloud/terraform-provider-hcloud/test/e2e/server"
	"github.com/hetznercloud/terraform-provider-hcloud/test/e2e/sshkey"
)

func TestVolumeResourceAttachment(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		var s hcloud.Server
		var s2 hcloud.Server
		var v hcloud.Volume
		tmplMan := testtemplate.Manager{}

		resSSHKey := sshkey.NewRData(t, "server-vol")
		resServer := &servertest.RData{
			Name:  "vol-attachment",
			Type:  teste2e.TestServerType,
			Image: teste2e.TestImage,
			Labels: map[string]string{
				"tf-test": fmt.Sprintf("tf-test-vol-attachment-%d", tmplMan.RandInt),
			},
			SSHKeys: []string{resSSHKey.TFID() + ".id"},
		}
		resServer.SetRName("server_attachment")

		resServer2 := &servertest.RData{
			Name:  "vol-attachment-2",
			Type:  teste2e.TestServerType,
			Image: teste2e.TestImage,
			Labels: map[string]string{
				"tf-test": fmt.Sprintf("tf-test-vol-attachment-%d", tmplMan.RandInt),
			},
			LocationName: fmt.Sprintf("${%s.location}", resServer.TFID()),
			SSHKeys:      []string{resSSHKey.TFID() + ".id"},
		}
		resServer2.SetRName("server2_attachment")

		resVolume := &RData{
			Name:         "volume-attachment",
			Size:         10,
			LocationName: fmt.Sprintf("${%s.location}", resServer.TFID()),
		}
		resVolume.SetRName("volume-attachment")

		res := &RDataAttachment{
			VolumeID: resVolume.TFID() + ".id",
			ServerID: resServer.TFID() + ".id",
		}

		resMove := &RDataAttachment{
			VolumeID: resVolume.TFID() + ".id",
			ServerID: resServer2.TFID() + ".id",
		}
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:                 teste2e.PreCheck(t),
			ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
			CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, servertest.ByID(t, &s)),
			Steps: []resource.TestStep{
				{
					// Create a new Volume attachment using the required values
					// only.
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_ssh_key", resSSHKey,
						"testdata/r/hcloud_server", resServer,
						"testdata/r/hcloud_server", resServer2,
						"testdata/r/hcloud_volume", resVolume,
						"testdata/r/hcloud_volume_attachment", res,
					),
					Check: resource.ComposeTestCheckFunc(
						testsupport.CheckResourceExists(resServer.TFID(), servertest.ByID(t, &s)),
						testsupport.CheckResourceExists(resVolume.TFID(), ByID(t, &v)),
					),
				},
				{
					// Try to import the newly created volume attachment
					ResourceName:      res.TFID(),
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: func(_ *terraform.State) (string, error) {
						return fmt.Sprintf("%d", v.ID), nil
					},
				},
				{
					// Move the Volume to another server using the
					// attachment.
					Config: tmplMan.Render(t,
						"testdata/r/hcloud_ssh_key", resSSHKey,
						"testdata/r/hcloud_server", resServer,
						"testdata/r/hcloud_server", resServer2,
						"testdata/r/hcloud_volume", resVolume,
						"testdata/r/hcloud_volume_attachment", resMove,
					),
					Check: resource.ComposeTestCheckFunc(
						testsupport.CheckResourceExists(resServer2.TFID(), servertest.ByID(t, &s2)),
						testsupport.CheckResourceExists(resVolume.TFID(), ByID(t, &v)),
					),
				},
			},
		})
	})
}
