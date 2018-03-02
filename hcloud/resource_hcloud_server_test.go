package hcloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

var (
	testAccSSHPublicKey string
)

func init() {
	resource.AddTestSweepers("hcloud_server", &resource.Sweeper{
		Name: "hcloud_server",
		F:    testSweepServers,
	})

	var err error
	testAccSSHPublicKey, _, err = acctest.RandSSHKeyPair("hcloud@ssh-acceptance-test")
	if err != nil {
		panic(fmt.Errorf("Cannot generate test SSH key pair: %s", err))
	}
}

func TestAccServer_Basic(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("hcloud_server.foobar", &server),
					testAccCheckServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "server_type", "cx11"),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "image", "debian-9"),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "status", "running"),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "location", "fsn1"),
				),
			},
		},
	})
}

func TestAccServer_Update(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("hcloud_server.foobar", &server),
					testAccCheckServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},

			{
				Config: testAccCheckServerConfig_RenameAndResize(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("hcloud_server.foobar", &server),
					testAccCheckServerRenamedAndResized(&server),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("baz-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "server_type", "cx21"),
				),
			},
		},
	})
}

func TestAccServer_UpdateUserData(t *testing.T) {
	var afterCreate, afterUpdate hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("hcloud_server.foobar", &afterCreate),
					testAccCheckServerAttributes(&afterCreate),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
				),
			},

			{
				Config: testAccCheckServerConfig_userdata_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("hcloud_server.foobar", &afterUpdate),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					testAccCheckServerRecreated(
						t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccServer_ISO(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerConfig_ISO(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("hcloud_server.foobar", &server),
					testAccCheckServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "iso", "coreos_stable_production.iso"),
				),
			},
		},
	})
}

func testAccCheckServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_server" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Try to find the Server
		_, _, err = client.Server.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error waiting for server (%s) to be destroyed: %v",
				rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckServerExists(n string, server *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		se, _, err := client.Server.GetByID(context.Background(), id)
		if err != nil {
			return err
		}

		if se == nil {
			return fmt.Errorf("Server not found")
		}

		*server = *se
		return nil
	}
}

func testAccCheckServerRenamedAndResized(server *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if server.ServerType.Name != "cx21" {
			return fmt.Errorf("Bad server.ServerType.Name: %s", server.ServerType.Name)
		}

		return nil
	}
}

func testAccCheckServerAttributes(server *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if server.Image.Name != "debian-9" {
			return fmt.Errorf("Bad server.Image.Name: %s", server.Image.Name)
		}

		if server.ServerType.Name != "cx11" {
			return fmt.Errorf("Bad server.ServerType.Name: %s", server.ServerType.Name)
		}

		if server.Datacenter.Location.Name != "fsn1" {
			return fmt.Errorf("Bad server.Datacenter.Location.Name: %s", server.Datacenter.Location.Name)
		}

		if server.BackupWindow != "22-02" {
			return fmt.Errorf("Bad server.BackupWindow: %s", server.BackupWindow)
		}

		return nil
	}
}

func testAccCheckServerRecreated(t *testing.T, before, after *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			t.Fatalf("Expected change of server IDs, but both were %v", before.ID)
		}
		return nil
	}
}

func testAccCheckServerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name          = "foo-%d"
  server_type   = "cx11"
  image         = "debian-9"
  datacenter    = "fsn1-dc8"
	user_data     = "stuff"
	backup_window = "22-02"
	ssh_keys  = ["${hcloud_ssh_key.foobar.id}"]
}`, rInt, testAccSSHPublicKey, rInt)
}

func testAccCheckServerConfig_ISO(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name        = "foo-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc8"
	backup_window = "22-02"
	iso         = "coreos_stable_production.iso"
	ssh_keys  = ["${hcloud_ssh_key.foobar.id}"]
}`, rInt, testAccSSHPublicKey, rInt)
}

func testAccCheckServerConfig_RenameAndResize(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name        = "baz-%d"
  server_type = "cx21"
	image       = "debian-9"
  datacenter  = "fsn1-dc8"
	ssh_keys  = ["${hcloud_ssh_key.foobar.id}"]
}
`, rInt, testAccSSHPublicKey, rInt)
}

func testAccCheckServerConfig_userdata_update(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name      = "foo-%d"
  server_type = "cx11"
	image       = "debian-9"
  datacenter  = "fsn1-dc8"
	user_data   = "updated stuff"
	ssh_keys  = ["${hcloud_ssh_key.foobar.id}"]
}
`, rInt, testAccSSHPublicKey, rInt)
}

func testSweepServers(region string) error {
	client, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	servers, err := client.Server.All(ctx)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Found %d servers to sweep", len(servers))

	for _, s := range servers {
		if strings.HasPrefix(s.Name, "terraform") {
			log.Printf("Deleting server %s", s.Name)
			if _, err := client.Server.Delete(ctx, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func createClient() (*hcloud.Client, error) {
	if os.Getenv("HCLOUD_TOKEN") == "" {
		return nil, fmt.Errorf("empty HCLOUD_TOKEN")
	}
	opts := []hcloud.ClientOption{
		hcloud.WithToken(os.Getenv("HCLOUD_TOKEN")),
	}
	return hcloud.NewClient(opts...), nil
}
