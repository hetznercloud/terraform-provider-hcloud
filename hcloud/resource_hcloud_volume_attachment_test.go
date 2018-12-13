package hcloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_volume_attachment", &resource.Sweeper{
		Name: "hcloud_volume_attachment",
		F:    testSweepVolumes,
	})
}

func TestAccHcloudVolumeAttachment_Create(t *testing.T) {
	var server hcloud.Server
	var volume hcloud.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckVolumeAttachmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume_attachment.foobar_attachment", &volume),
					testAccHcloudCheckVolumeAttachmentVolume("hcloud_volume_attachment.foobar_attachment", &volume),
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					testAccHcloudCheckVolumeAttachmentServer("hcloud_volume_attachment.foobar_attachment", &server),
				),
			},
		},
	})
}

func TestAccHcloudVolumeAttachment_CreateMany(t *testing.T) {
	var server hcloud.Server
	var volume hcloud.Volume
	var volume2 hcloud.Volume
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckMultipleVolumeAttachmentConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckVolumeExists("hcloud_volume_attachment.foobar_attachment", &volume),
					testAccHcloudCheckVolumeAttachmentVolume("hcloud_volume_attachment.foobar_attachment", &volume),
					testAccHcloudCheckVolumeExists("hcloud_volume_attachment.foobar_attachment2", &volume2),
					testAccHcloudCheckVolumeAttachmentVolume("hcloud_volume_attachment.foobar_attachment2", &volume2),
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					testAccHcloudCheckVolumeAttachmentServer("hcloud_volume_attachment.foobar_attachment", &server),
				),
			},
		},
	})
}

func testAccHcloudCheckVolumeAttachmentVolume(n string, volume *hcloud.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		id := rs.Primary.Attributes["volume_id"]

		if id != strconv.Itoa(volume.ID) {
			return fmt.Errorf("Volume Attachment volume id is not valid: %v", id)
		}

		return nil
	}
}

func testAccHcloudCheckVolumeAttachmentServer(n string, server *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		id := rs.Primary.Attributes["server_id"]

		if id != strconv.Itoa(server.ID) {
			return fmt.Errorf("Volume Attachment Server id is not valid: %v", id)
		}

		return nil
	}
}

func testAccHcloudCheckVolumeAttachmentConfig(serverID int) string {
	return fmt.Sprintf(`
resource "hcloud_volume_attachment" "foobar_attachment" {
  volume_id = "${hcloud_volume.foobar_volume.id}"
  server_id = "${hcloud_server.foobar.id}"
  automount = true
}

resource "hcloud_server" "foobar" {
  name        = "foo-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "nbg1-dc3"
}

resource "hcloud_volume" "foobar_volume" {
  size     = 10
  location = "nbg1"
  name     = "foo-volume-%d"
}
`, serverID, serverID)
}

func testAccHcloudCheckMultipleVolumeAttachmentConfig(serverID int) string {
	return fmt.Sprintf(`
resource "hcloud_volume_attachment" "foobar_attachment" {
  volume_id = "${hcloud_volume.foobar_volume.id}"
  server_id = "${hcloud_server.foobar.id}"
}

resource "hcloud_volume_attachment" "foobar_attachment2" {
  volume_id = "${hcloud_volume.foobar_volume2.id}"
  server_id = "${hcloud_server.foobar.id}"
}

resource "hcloud_server" "foobar" {
  name        = "foo-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "nbg1-dc3"
}

resource "hcloud_volume" "foobar_volume" {
  size     = 10
  location = "nbg1"
  name     = "foo-volume-%d"
}
resource "hcloud_volume" "foobar_volume2" {
  size     = 10
  location = "nbg1"
  name     = "foo-volume-2-%d"
}
`, serverID, serverID, serverID)
}
