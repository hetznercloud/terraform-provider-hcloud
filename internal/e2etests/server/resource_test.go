package server_test

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/e2etests"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/placementgroup"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestServerResource_Basic(t *testing.T) {
	var s hcloud.Server

	sk := sshkey.NewRData(t, "server-basic")
	res := &server.RData{
		Name:    "server-basic",
		Type:    e2etests.TestServerType,
		Image:   e2etests.TestImage,
		SSHKeys: []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-basic")
	resRenamed := &server.RData{Name: res.Name + "-renamed", Type: res.Type, Image: res.Image}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-basic--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
				),
			},
			{
				// Try to import the newly created Server
				ResourceName:      res.TFID(),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ssh_keys", "user_data", "keep_disk"},
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields and renaming the Server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.TFID(), "name",
						fmt.Sprintf("server-basic-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(resRenamed.TFID(), "image", res.Image),
				),
			},
		},
	})
}

func TestServerResource_Resize(t *testing.T) {
	var s hcloud.Server

	sk := sshkey.NewRData(t, "server-resize")
	res := &server.RData{
		Name:    "server-resize",
		Type:    e2etests.TestServerType,
		Image:   e2etests.TestImage,
		SSHKeys: []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-resize")
	resResized := &server.RData{Name: res.Name, Type: "cx21", Image: res.Image, KeepDisk: true}
	resResized.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
				),
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields and renaming the Server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", resResized,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resResized.TFID(), "name",
						fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resResized.TFID(), "server_type", resResized.Type),
					resource.TestCheckResourceAttr(resResized.TFID(), "image", res.Image),
				),
			},
		},
	})
}

func TestServerResource_ChangeUserData(t *testing.T) {
	var s, s2 hcloud.Server

	sk := sshkey.NewRData(t, "server-userdata")
	res := &server.RData{
		Name:     "server-userdata",
		Type:     e2etests.TestServerType,
		Image:    e2etests.TestImage,
		UserData: "stuff",
		SSHKeys:  []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-userdata")
	resChangedUserdata := &server.RData{Name: res.Name, Type: res.Type, Image: res.Image, UserData: "updated stuff"}
	resChangedUserdata.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(res.TFID(), "user_data", userDataHashSum(res.UserData)),
				),
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields and renaming the Server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", resChangedUserdata,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s2)),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "name",
						fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "user_data", userDataHashSum(resChangedUserdata.UserData)),
					testsupport.LiftTCF(isRecreated(&s2, &s)),
				),
			},
		},
	})
}

func TestServerResource_ISO(t *testing.T) {
	var s hcloud.Server

	sk := sshkey.NewRData(t, "server-iso")
	res := &server.RData{
		Name:     "server-iso",
		Type:     e2etests.TestServerType,
		Image:    e2etests.TestImage,
		UserData: "stuff",
		ISO:      "3500",
		SSHKeys:  []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-iso")
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-iso--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(res.TFID(), "iso", res.ISO),
				),
			},
		},
	})
}

func TestServerResource_DirectAttachToNetwork(t *testing.T) {
	var (
		nw hcloud.Network
		s  hcloud.Server

		// Helper functions to modify the test data. Those functions modify
		// the passed in server on purpose. Calling them once to change the
		// respective value es enough.
		updateIP = func(d *server.RData, ip string) *server.RData {
			d.Network.IP = ip
			return d
		}
		updateAliasIPs = func(d *server.RData, ips ...string) *server.RData {
			d.Network.AliasIPs = ips
			return d
		}
	)

	sk := sshkey.NewRData(t, "server-iso")
	nwRes := &network.RData{
		Name:    "test-network",
		IPRange: "10.0.0.0/16",
	}
	nwRes.SetRName("test-network")
	snwRes := &network.RDataSubnet{
		Type:        "cloud",
		NetworkID:   nwRes.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	snwRes.SetRName("test-network-subnet")
	sRes := &server.RData{
		Name:         "server-direct-attach",
		Type:         e2etests.TestServerType,
		LocationName: e2etests.TestLocationName,
		Image:        e2etests.TestImage,
		SSHKeys:      []string{sk.TFID() + ".id"},
	}
	sRes.SetRName(sRes.Name)

	sResWithNet := &server.RData{
		Name:         sRes.Name,
		Type:         sRes.Type,
		LocationName: sRes.LocationName,
		Image:        sRes.Image,
		SSHKeys:      sRes.SSHKeys,
		Network: server.RDataInlineNetwork{
			NetworkID: nwRes.TFID() + ".id",
			IP:        "10.0.1.5",
			AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
		},
		DependsOn: []string{snwRes.TFID()},
	}
	sResWithNet.SetRName(sResWithNet.Name)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a new server and directly attach it to a network.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_server", sResWithNet,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nwRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(sResWithNet.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw, "10.0.1.5", "10.0.1.6", "10.0.1.7")),
					resource.TestCheckResourceAttr(sResWithNet.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(sResWithNet.TFID(), "network.0.ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(sResWithNet.TFID(), "network.0.alias_ips.#", "2"),
				),
			},
			{
				// Change the IP of the server
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_server", updateIP(sResWithNet, "10.0.1.4"),
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nwRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(sResWithNet.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw, "10.0.1.4", "10.0.1.6", "10.0.1.7")),
				),
			},
			{
				// Change the AliasIPs of the server
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_server", updateAliasIPs(sResWithNet, "10.0.1.5", "10.0.1.7"),
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nwRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(sResWithNet.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw, "10.0.1.4", "10.0.1.5", "10.0.1.7")),
				),
			},
			{
				// Detach the server from the network.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_server", sRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(sRes.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(func() error {
						t.Log("Checking if server has no private network")
						assert.Empty(t, s.PrivateNet)
						return nil
					}),
				),
			},
		},
	})
}

func TestServerResource_Firewalls(t *testing.T) {
	var s hcloud.Server

	fw := firewall.NewRData(t, "server-test", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "icmp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
		},
	}, nil)
	fw2 := firewall.NewRData(t, "server-test-2", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "tcp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "1-65535",
		},
	}, nil)
	res := &server.RData{
		Name:        "server-firewall",
		Type:        e2etests.TestServerType,
		Image:       e2etests.TestImage,
		FirewallIDs: []string{fw.TFID() + ".id"},
	}
	res.SetRName("server-firewall")
	res2 := &server.RData{
		Name:        "server-firewall",
		Type:        e2etests.TestServerType,
		Image:       e2etests.TestImage,
		FirewallIDs: []string{fw2.TFID() + ".id"},
	}
	res2.SetRName(res.RName())
	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_firewall", fw,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(res.TFID(), "firewall_ids.#", "1"),
				),
			},
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_firewall", fw,
					"testdata/r/hcloud_firewall", fw2,
					"testdata/r/hcloud_server", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
					resource.TestCheckResourceAttr(res.TFID(), "name",
						fmt.Sprintf("server-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(res.TFID(), "firewall_ids.#", "1"),
				),
			},
		},
	})
}

func TestServerResource_PlacementGroup(t *testing.T) {
	var (
		pg  hcloud.PlacementGroup
		srv hcloud.Server
	)

	pgRes := placementgroup.NewRData(t, "server-test", "spread")

	srvRes := &server.RData{
		Name:             "server-placement-group",
		Type:             e2etests.TestServerType,
		Image:            e2etests.TestImage,
		PlacementGroupID: pgRes.TFID() + ".id",
	}
	srvRes.SetRName("server-placement-group")

	tmplMan := testtemplate.Manager{}

	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", pgRes,
					"testdata/r/hcloud_server", srvRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(srvRes.TFID(), server.ByID(t, &srv)),
					testsupport.CheckResourceExists(pgRes.TFID(), placementgroup.ByID(t, &pg)),
					resource.TestCheckResourceAttr(srvRes.TFID(), "name",
						fmt.Sprintf("server-placement-group--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(srvRes.TFID(), "server_type", srvRes.Type),
					resource.TestCheckResourceAttr(srvRes.TFID(), "image", srvRes.Image),
					testsupport.CheckResourceAttrFunc(srvRes.TFID(), "placement_group_id", func() string {
						return strconv.Itoa(pg.ID)
					}),
				),
			},
		},
	})
}

func TestServerResource_Protection(t *testing.T) {
	var (
		srv hcloud.Server

		updateProtection = func(d *server.RData, protection bool) *server.RData {
			d.DeleteProtection = protection
			d.RebuildProtection = protection
			return d
		}
	)

	srvRes := &server.RData{
		Name:              "server-protection",
		Type:              e2etests.TestServerType,
		Image:             e2etests.TestImage,
		DeleteProtection:  true,
		RebuildProtection: true,
	}
	srvRes.SetRName("server-protection")

	tmplMan := testtemplate.Manager{}

	resource.Test(t, resource.TestCase{
		PreCheck:     e2etests.PreCheck(t),
		Providers:    e2etests.Providers(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvRes,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(srvRes.TFID(), server.ByID(t, &srv)),
					resource.TestCheckResourceAttr(srvRes.TFID(), "name",
						fmt.Sprintf("server-protection--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(srvRes.TFID(), "server_type", srvRes.Type),
					resource.TestCheckResourceAttr(srvRes.TFID(), "image", srvRes.Image),
					resource.TestCheckResourceAttr(srvRes.TFID(), "delete_protection", fmt.Sprintf("%t", srvRes.DeleteProtection)),
					resource.TestCheckResourceAttr(srvRes.TFID(), "rebuild_protection", fmt.Sprintf("%t", srvRes.RebuildProtection)),
				),
			},
			{
				// Update delete protection
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", updateProtection(srvRes, false),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(srvRes.TFID(), "delete_protection", fmt.Sprintf("%t", srvRes.DeleteProtection)),
					resource.TestCheckResourceAttr(srvRes.TFID(), "rebuild_protection", fmt.Sprintf("%t", srvRes.RebuildProtection)),
				),
			},
		},
	})
}

func isRecreated(new, old *hcloud.Server) func() error {
	return func() error {
		if new.ID == old.ID {
			return fmt.Errorf("new server is the same as server cert %d", old.ID)
		}
		return nil
	}
}

func userDataHashSum(userData string) string {
	sum := sha1.Sum([]byte(userData))
	return base64.StdEncoding.EncodeToString(sum[:])
}
