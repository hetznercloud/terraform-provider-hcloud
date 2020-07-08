package hcloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

var (
	testAccSSHPublicKey string
	testHcloudISOID     = "3500"
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

func TestAccHcloudServer_Basic(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					testAccHcloudCheckServerAttributes(&server),
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
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "backups", "true"),
				),
			},
		},
	})
}

func TestAccHcloudServer_Update(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					testAccHcloudCheckServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "backups", "true"),
				),
			},

			{
				Config: testAccHcloudCheckServerConfig_RenameAndResize(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					testAccHcloudCheckServerRenamedAndResized(&server),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("baz-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "server_type", "cx21"),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "backups", "false"),
				),
			},
		},
	})
}

func TestAccHcloudServer_UpdateUserData(t *testing.T) {
	var afterCreate, afterUpdate hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckServerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.foobar", &afterCreate),
					testAccHcloudCheckServerAttributes(&afterCreate),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "user_data", userDataHashSum("stuff")),
				),
			},

			{
				Config: testAccHcloudCheckServerConfig_userdata_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.foobar", &afterUpdate),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "user_data", userDataHashSum("updated stuff")),
					testAccHcloudCheckServerRecreated(
						t, &afterCreate, &afterUpdate),
				),
			},
		},
	})
}

func TestAccHcloudServer_ISOID(t *testing.T) {
	var server hcloud.Server
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccHcloudPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccHcloudCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHcloudCheckServerConfig_ISOID(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccHcloudCheckServerExists("hcloud_server.foobar", &server),
					testAccHcloudCheckServerAttributes(&server),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "name", fmt.Sprintf("foo-%d", rInt)),
					resource.TestCheckResourceAttr(
						"hcloud_server.foobar", "iso", testHcloudISOID),
				),
			},
		},
	})
}

func testAccHcloudCheckServerDestroy(s *terraform.State) error {
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
		var server *hcloud.Server
		server, _, err = client.Server.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error checking if server (%s) is deleted: %v",
				rs.Primary.ID, err)
		}
		if server != nil {
			return fmt.Errorf("Server (%s) has not been deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccHcloudCheckServerExists(n string, server *hcloud.Server) resource.TestCheckFunc {
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

func testAccHcloudCheckServerRenamedAndResized(server *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if server.ServerType.Name != "cx21" {
			return fmt.Errorf("Bad server.ServerType.Name: %s", server.ServerType.Name)
		}

		return nil
	}
}

func testAccHcloudCheckServerAttributes(server *hcloud.Server) resource.TestCheckFunc {
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

		if server.BackupWindow == "" {
			return fmt.Errorf("Bad server.BackupWindow: %s", server.BackupWindow)
		}

		return nil
	}
}

func testAccHcloudCheckServerRecreated(t *testing.T, before, after *hcloud.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ID == after.ID {
			t.Fatalf("Expected change of server IDs, but both were %v", before.ID)
		}
		return nil
	}
}

func testAccHcloudCheckServerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name          = "foo-%d"
  server_type   = "cx11"
  image         = "debian-9"
  datacenter    = "fsn1-dc14"
  user_data     = "stuff"
  backups       = true
  ssh_keys      = ["${hcloud_ssh_key.foobar.id}"]
}`, rInt, testAccSSHPublicKey, rInt)
}

func testAccHcloudCheckServerConfig_ISOID(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name        = "foo-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc14"
  backups     = true
  iso         = "%s"
  ssh_keys    = ["${hcloud_ssh_key.foobar.id}"]
}`, rInt, testAccSSHPublicKey, rInt, testHcloudISOID)
}

func testAccHcloudCheckServerConfig_RenameAndResize(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name        = "baz-%d"
  server_type = "cx21"
  image       = "debian-9"
  datacenter  = "fsn1-dc14"
  ssh_keys    = ["${hcloud_ssh_key.foobar.id}"]
  backups     =  false
}
`, rInt, testAccSSHPublicKey, rInt)
}

func testAccHcloudCheckServerConfig_userdata_update(rInt int) string {
	return fmt.Sprintf(`
resource "hcloud_ssh_key" "foobar" {
  name       = "foobar-%d"
  public_key = "%s"
}
resource "hcloud_server" "foobar" {
  name      = "foo-%d"
  server_type = "cx11"
  image       = "debian-9"
  datacenter  = "fsn1-dc14"
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

// TODO remove me
// Deprecated: use testsupport.CreateClient instead
func createClient() (*hcloud.Client, error) {
	if os.Getenv("HCLOUD_TOKEN") == "" {
		return nil, fmt.Errorf("empty HCLOUD_TOKEN")
	}
	opts := []hcloud.ClientOption{
		hcloud.WithToken(os.Getenv("HCLOUD_TOKEN")),
	}
	return hcloud.NewClient(opts...), nil
}
