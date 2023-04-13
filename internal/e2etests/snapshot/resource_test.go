package snapshot

import (
	"fmt"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/snapshot"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestSnapshotResource_Basic(t *testing.T) {
	var s hcloud.Image
	tmplMan := testtemplate.Manager{}

	sk := sshkey.NewRData(t, "snapshot-basic")
	resServer := &server.RData{
		Name:  "snapshot-test",
		Type:  e2etests.TestServerType,
		Image: e2etests.TestImage,
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-snapshot-%d", tmplMan.RandInt),
		},
		SSHKeys: []string{sk.TFID() + ".id"},
	}
	resServer.SetRName("server-snapshot")
	res := &snapshot.RData{
		Description: "snapshot-basic",
		ServerID:    resServer.TFID() + ".id",
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-snapshot-%d", tmplMan.RandInt),
		},
	}
	res.SetRName("snapshot-basic")
	resRenamed := &snapshot.RData{
		Description: "snapshot-basic-changed",
		ServerID:    resServer.TFID() + ".id",
		Labels: map[string]string{
			"tf-test": fmt.Sprintf("tf-test-fip-assignment-%d", tmplMan.RandInt),
		}}
	resRenamed.SetRName("snapshot-basic")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(snapshot.ResourceType, snapshot.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Snapshot using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_snapshot", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), snapshot.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "description", "snapshot-basic"),
				),
			},
			{
				// Try to import the newly created Snapshot
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the Snapshot created in the previous step by
				// setting all optional fields and renaming the Snapshot.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", resServer,
					"testdata/r/hcloud_snapshot", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), snapshot.ByID(t, &s)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "description", "snapshot-basic-changed"),
				),
			},
		},
	})
}
