package server_test

import (
	"context"
	"crypto/sha1" // nolint: gosec
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/firewall"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/image"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/placementgroup"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

func TestAccServerResource(t *testing.T) {
	var s hcloud.Server

	sk := sshkey.NewRData(t, "server-basic")
	res := &server.RData{
		Name:    "server-basic",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-basic")
	resRenamed := &server.RData{Name: res.Name + "-renamed", Type: res.Type, Image: res.Image}
	resRenamed.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
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
					"ssh_keys", "user_data", "keep_disk", "ignore_remote_firewall_ids", "allow_deprecated_images", "shutdown_before_deletion",
				},
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

func TestAccServerResource_UnavailableServerType(t *testing.T) {
	var s hcloud.Server

	res := &server.RData{
		Name:  "unavailable-server-type",
		Type:  "1",
		Image: teste2e.TestImage,
	}
	res.SetRName(res.Name)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res,
				),
				ExpectError: regexp.MustCompile(`Server Type "cx11" is unavailable in all locations and can no longer be ordered.`),
			},
		},
	})
}

func TestAccServerResource_ImageID(t *testing.T) {
	var s hcloud.Server

	sk := sshkey.NewRData(t, "server-image-id")
	img := &image.DData{
		ImageName:    teste2e.TestImage,
		Architecture: hcloud.ArchitectureX86,
	}
	img.SetRName("server-image-id")
	res := &server.RData{
		Name:    "server-image-id",
		Type:    teste2e.TestServerType,
		Image:   fmt.Sprintf("${%s.id}", img.TFID()),
		SSHKeys: []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-image-id")
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/d/hcloud_image", img,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &s)),
				),
			},
		},
	})
}

func TestAccServerResource_Resize(t *testing.T) {
	var s hcloud.Server

	sk := sshkey.NewRData(t, "server-resize")
	res := &server.RData{
		Name:    "server-resize",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-resize")
	resResized := &server.RData{
		Name:     res.Name,
		Type:     teste2e.TestServerTypeUpgrade,
		KeepDisk: true,
		Image:    res.Image,
		SSHKeys:  res.SSHKeys,
	}
	resResized.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
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
					resource.TestCheckResourceAttr(res.TFID(), "name", fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
				),
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", resResized,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resResized.TFID(), "name", fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resResized.TFID(), "server_type", resResized.Type),
					resource.TestCheckResourceAttr(resResized.TFID(), "image", resResized.Image),
				),
			},
		},
	})
}

func TestAccServerResource_ChangeUserData(t *testing.T) {
	var s, s2 hcloud.Server

	sk := sshkey.NewRData(t, "server-userdata")
	res := &server.RData{
		Name:     "server-userdata",
		Type:     teste2e.TestServerType,
		Image:    teste2e.TestImage,
		UserData: "stuff",
		SSHKeys:  []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-userdata")
	resChangedUserdata := &server.RData{Name: res.Name, Type: res.Type, Image: res.Image, UserData: "updated stuff"}
	resChangedUserdata.SetRName(res.Name)
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
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
					resource.TestCheckResourceAttr(res.TFID(), "name", fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(res.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(res.TFID(), "user_data", userDataHashSum(res.UserData+"\n")),
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
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "name", fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "server_type", res.Type),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "image", res.Image),
					resource.TestCheckResourceAttr(resChangedUserdata.TFID(), "user_data", userDataHashSum(resChangedUserdata.UserData+"\n")),
					testsupport.LiftTCF(isRecreated(&s2, &s)),
				),
			},
		},
	})
}

func TestAccServerResource_ISO(t *testing.T) {
	var s hcloud.Server

	sk := sshkey.NewRData(t, "server-iso")
	res := &server.RData{
		Name:     "server-iso",
		Type:     teste2e.TestServerType,
		Image:    teste2e.TestImage,
		UserData: "stuff",
		ISO:      "8637", // Windows Server 2022 English
		SSHKeys:  []string{sk.TFID() + ".id"},
	}
	res.SetRName("server-iso")
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
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

func TestAccServerResource_DirectAttachToNetwork(t *testing.T) {
	var (
		nw  hcloud.Network
		nw2 hcloud.Network
		s   hcloud.Server

		// Helper functions to modify the test data. Those functions modify
		// the passed in server on purpose. Calling them once to change the
		// respective value is enough.
		updateIP = func(d *server.RData, networkID string, ip string) *server.RData {
			for i := range d.Networks {
				if d.Networks[i].NetworkID == networkID {
					d.Networks[i].IP = ip
				}
			}
			return d
		}
		updateAliasIPs = func(d *server.RData, networkID string, ips ...string) *server.RData {
			for i := range d.Networks {
				if d.Networks[i].NetworkID == networkID {
					d.Networks[i].AliasIPs = ips
				}
			}
			return d
		}

		addNetwork = func(d *server.RData, network server.RDataInlineNetwork) *server.RData {
			d.Networks = append(d.Networks, network)
			return d
		}
	)

	sk := sshkey.NewRData(t, "server-direct-attach-network")

	// Network 1
	nwRes := &network.RData{
		Name:    "test-network-1",
		IPRange: "10.0.0.0/16",
	}
	nwRes.SetRName("test-network-1")
	snwRes := &network.RDataSubnet{
		Type:        "cloud",
		NetworkID:   nwRes.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	snwRes.SetRName("test-network-subnet-1")

	// Network 2
	nw2Res := &network.RData{
		Name:    "test-network-2",
		IPRange: "10.1.0.0/16",
	}
	nw2Res.SetRName("test-network-2")
	snw2Res := &network.RDataSubnet{
		Type:        "cloud",
		NetworkID:   nw2Res.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.1.1.0/24",
	}
	snw2Res.SetRName("test-network-subnet-2")

	sRes := &server.RData{
		Name:         "server-direct-attach",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{sk.TFID() + ".id"},
	}
	sRes.SetRName(sRes.Name)

	sResWithNet := &server.RData{
		Name:         sRes.Name,
		Type:         sRes.Type,
		LocationName: sRes.LocationName,
		Image:        sRes.Image,
		SSHKeys:      sRes.SSHKeys,
		Networks: []server.RDataInlineNetwork{{
			NetworkID: nwRes.TFID() + ".id",
			IP:        "10.0.1.5",
			AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
		}},
		DependsOn: []string{snwRes.TFID()},
	}
	sResWithNet.SetRName(sResWithNet.Name)

	sResWithTwoNets := &server.RData{
		Name:         sRes.Name,
		Type:         sRes.Type,
		LocationName: sRes.LocationName,
		Image:        sRes.Image,
		SSHKeys:      sRes.SSHKeys,
		Networks: []server.RDataInlineNetwork{
			{
				NetworkID: nwRes.TFID() + ".id",
				IP:        "10.0.1.5",
				AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
			},
			{
				NetworkID: nw2Res.TFID() + ".id",
				IP:        "10.1.1.5",
				AliasIPs:  []string{"10.1.1.6", "10.1.1.7"},
			},
		},
		DependsOn: []string{snwRes.TFID(), snw2Res.TFID()},
	}
	sResWithTwoNets.SetRName(sResWithTwoNets.Name)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
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
					"testdata/r/hcloud_server", updateIP(sResWithNet, nwRes.TFID()+".id", "10.0.1.4"),
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
					"testdata/r/hcloud_server", updateAliasIPs(sResWithNet, nwRes.TFID()+".id", "10.0.1.5", "10.0.1.7"),
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
			{
				// Fail when using conflicting networks
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_server", addNetwork(sResWithNet, server.RDataInlineNetwork{
						NetworkID: nwRes.TFID() + ".id",
						IP:        "10.0.1.8",
					}),
				),
				ExpectError: regexp.MustCompile(`server is only allowed to be attached to each network once: \d+`),
			},

			{
				// Remove networks and test to attach to two new networks at the same time
				// Fail when using conflicting networks
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
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

			{
				// Continuation of above test
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_network", nw2Res,
					"testdata/r/hcloud_network_subnet", snw2Res,
					"testdata/r/hcloud_server", sResWithTwoNets,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(sRes.TFID(), server.ByID(t, &s)),
					testsupport.CheckResourceExists(nwRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(nw2Res.TFID(), network.ByID(t, &nw2)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw, "10.0.1.5", "10.0.1.6", "10.0.1.7")),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw2, "10.1.1.5", "10.1.1.6", "10.1.1.7")),
				),
			},
		},
	})
}

func TestAccServerResource_PrimaryIPNetworkTests(t *testing.T) {
	var (
		nw hcloud.Network
		s  hcloud.Server
		p  hcloud.PrimaryIP
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

	primaryIPv4Res := &primaryip.RData{
		Name:         "primaryip-v4-test",
		Type:         "ipv4",
		Labels:       nil,
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPv4Res.SetRName("primary-ip-v4")

	primaryIPv6Res := &primaryip.RData{
		Name:         "primaryip-v6-test",
		Type:         "ipv6",
		Labels:       nil,
		Datacenter:   teste2e.TestDataCenter,
		AssigneeType: "server",
		AutoDelete:   false,
	}
	primaryIPv6Res.SetRName("primary-ip-v6")

	sResWithNetAndPublicNet := &server.RData{
		Name:       "server-primaryIP-network-test",
		Type:       teste2e.TestServerType,
		Datacenter: teste2e.TestDataCenter,
		Image:      teste2e.TestImage,
		SSHKeys:    []string{sk.TFID() + ".id"},
		Networks: []server.RDataInlineNetwork{{
			NetworkID: nwRes.TFID() + ".id",
			IP:        "10.0.1.5",
			AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
		}},
		PublicNet: map[string]interface{}{
			"ipv4_enabled": true,
			"ipv6_enabled": true,
		},
		DependsOn: []string{snwRes.TFID()},
	}
	sResWithNetAndPublicNet.SetRName(sResWithNetAndPublicNet.Name)

	sResWithoutPublicNet := &server.RData{
		Name:       sResWithNetAndPublicNet.Name,
		Type:       sResWithNetAndPublicNet.Type,
		Datacenter: sResWithNetAndPublicNet.Datacenter,
		Image:      sResWithNetAndPublicNet.Image,
		SSHKeys:    sResWithNetAndPublicNet.SSHKeys,
		Networks:   sResWithNetAndPublicNet.Networks,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": false,
			"ipv6_enabled": false,
		},
		DependsOn: sResWithNetAndPublicNet.DependsOn,
	}
	sResWithoutPublicNet.SetRName(sResWithoutPublicNet.Name)

	sResWithPrimaryIP := &server.RData{
		Name:       sResWithoutPublicNet.Name,
		Type:       sResWithoutPublicNet.Type,
		Datacenter: sResWithoutPublicNet.Datacenter,
		Image:      sResWithoutPublicNet.Image,
		SSHKeys:    sResWithoutPublicNet.SSHKeys,
		Networks:   sResWithoutPublicNet.Networks,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": true,
			"ipv4":         primaryIPv4Res.TFID() + ".id",
			"ipv6_enabled": false,
		},
		DependsOn: sResWithoutPublicNet.DependsOn,
	}

	sResWithPrimaryIP.SetRName(sResWithPrimaryIP.Name)

	sResWithTwoPrimaryIPs := &server.RData{
		Name:       sResWithPrimaryIP.Name,
		Type:       sResWithPrimaryIP.Type,
		Datacenter: sResWithPrimaryIP.Datacenter,
		Image:      sResWithPrimaryIP.Image,
		SSHKeys:    sResWithPrimaryIP.SSHKeys,
		Networks:   sResWithPrimaryIP.Networks,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": true,
			"ipv4":         primaryIPv4Res.TFID() + ".id",
			"ipv6_enabled": true,
		},
		DependsOn: sResWithoutPublicNet.DependsOn,
	}

	sResWithTwoPrimaryIPs.SetRName(sResWithTwoPrimaryIPs.Name)

	sResWithNoPublicNet := &server.RData{
		Name:       sResWithTwoPrimaryIPs.Name,
		Type:       sResWithTwoPrimaryIPs.Type,
		Datacenter: sResWithTwoPrimaryIPs.Datacenter,
		Image:      sResWithTwoPrimaryIPs.Image,
		SSHKeys:    sResWithTwoPrimaryIPs.SSHKeys,
		Networks:   sResWithTwoPrimaryIPs.Networks,
		DependsOn:  sResWithTwoPrimaryIPs.DependsOn,
	}

	sResWithNoPublicNet.SetRName(sResWithNoPublicNet.Name)

	sResWithOnlyIPv6 := &server.RData{
		Name:       sResWithNoPublicNet.Name,
		Type:       sResWithNoPublicNet.Type,
		Datacenter: sResWithNoPublicNet.Datacenter,
		Image:      sResWithNoPublicNet.Image,
		SSHKeys:    sResWithNoPublicNet.SSHKeys,
		Networks:   sResWithNoPublicNet.Networks,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": false,
			"ipv6_enabled": true,
			"ipv6":         primaryIPv6Res.TFID() + ".id",
		},
		DependsOn: sResWithNoPublicNet.DependsOn,
	}

	sResWithOnlyIPv6.SetRName(sResWithOnlyIPv6.Name)

	sResWithOnlyIPv6AutoGenerated := &server.RData{
		Name:       sResWithOnlyIPv6.Name,
		Type:       sResWithOnlyIPv6.Type,
		Datacenter: sResWithOnlyIPv6.Datacenter,
		Image:      sResWithOnlyIPv6.Image,
		SSHKeys:    sResWithOnlyIPv6.SSHKeys,
		Networks:   sResWithOnlyIPv6.Networks,
		PublicNet: map[string]interface{}{
			"ipv4_enabled": false,
			"ipv6_enabled": true,
		},
		DependsOn: sResWithOnlyIPv6.DependsOn,
	}

	sResWithOnlyIPv6AutoGenerated.SetRName(sResWithOnlyIPv6AutoGenerated.Name)

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				// Create a new server with unmanaged primary IPs + network
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_server", sResWithNetAndPublicNet,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nwRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(sResWithNetAndPublicNet.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw, "10.0.1.5", "10.0.1.6", "10.0.1.7")),
					resource.TestCheckResourceAttr(sResWithNetAndPublicNet.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(sResWithNetAndPublicNet.TFID(), "network.0.ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(sResWithNetAndPublicNet.TFID(), "network.0.alias_ips.#", "2"),
					testsupport.LiftTCF(func() error {
						assert.NotEqual(t, int64(0), s.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), s.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				// Primary IPs getting removed
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_server", sResWithoutPublicNet,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nwRes.TFID(), network.ByID(t, &nw)),
					testsupport.CheckResourceExists(sResWithoutPublicNet.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(hasServerNetwork(t, &s, &nw, "10.0.1.5", "10.0.1.6", "10.0.1.7")),
					resource.TestCheckResourceAttr(sResWithoutPublicNet.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(sResWithoutPublicNet.TFID(), "network.0.ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(sResWithoutPublicNet.TFID(), "network.0.alias_ips.#", "2"),
					testsupport.LiftTCF(func() error {
						assert.Nil(t, s.PublicNet.IPv4.IP)
						assert.Nil(t, s.PublicNet.IPv6.IP)
						return nil
					}),
				),
			},
			{
				// Add ipv4 via ID
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_primary_ip", primaryIPv4Res,
					"testdata/r/hcloud_server", sResWithPrimaryIP,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(primaryIPv4Res.TFID(), primaryip.ByID(t, &p)),
					testsupport.CheckResourceExists(sResWithPrimaryIP.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(func() error {
						assert.Equal(t, p.AssigneeID, s.ID)
						assert.Equal(t, p.ID, s.PublicNet.IPv4.ID)
						assert.Equal(t, int64(0), s.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				// Add ipv6 but auto generated (only ipv6_enabled = true, without an ID)
				// now ipv4 is a TF resource and ipv6 is auto generated
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_primary_ip", primaryIPv4Res,
					"testdata/r/hcloud_server", sResWithTwoPrimaryIPs,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(primaryIPv4Res.TFID(), primaryip.ByID(t, &p)),
					testsupport.CheckResourceExists(sResWithPrimaryIP.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(func() error {
						assert.Equal(t, p.AssigneeID, s.ID)
						assert.Equal(t, s.PublicNet.IPv4.ID, p.ID)
						assert.NotEqual(t, int64(0), s.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				// Remove public net, so attached ipv4 gets unattached + an ipv4 should be auto generated
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_primary_ip", primaryIPv4Res,
					"testdata/r/hcloud_server", sResWithNoPublicNet,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(primaryIPv4Res.TFID(), primaryip.ByID(t, &p)),
					testsupport.CheckResourceExists(sResWithPrimaryIP.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(func() error {
						assert.NotEqual(t, p.ID, s.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), s.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), s.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				// should remove auto generated ipv4 / 6 + attach managed ipv6
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_primary_ip", primaryIPv6Res,
					"testdata/r/hcloud_server", sResWithOnlyIPv6,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(primaryIPv6Res.TFID(), primaryip.ByID(t, &p)),
					testsupport.CheckResourceExists(sResWithOnlyIPv6.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(func() error {
						assert.Equal(t, p.ID, s.PublicNet.IPv6.ID)
						assert.Equal(t, int64(0), s.PublicNet.IPv4.ID)
						return nil
					}),
				),
			},
			{
				// should remove attached ipv6 and auto generate an ipv6
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_network", nwRes,
					"testdata/r/hcloud_network_subnet", snwRes,
					"testdata/r/hcloud_primary_ip", primaryIPv6Res,
					"testdata/r/hcloud_server", sResWithOnlyIPv6AutoGenerated,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(primaryIPv6Res.TFID(), primaryip.ByID(t, &p)),
					testsupport.CheckResourceExists(sResWithOnlyIPv6AutoGenerated.TFID(), server.ByID(t, &s)),
					testsupport.LiftTCF(func() error {
						assert.NotEqual(t, p.ID, s.PublicNet.IPv4.ID)
						assert.Equal(t, int64(0), s.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), s.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
		},
	})
}

func TestAccServerResource_PrivateNetworkBastion(t *testing.T) {
	name := "server-private-network-bastion"

	sshKeyRes := sshkey.NewRData(t, name)

	networkRes := &network.RData{Name: name, IPRange: "10.0.0.0/16"}
	networkRes.SetRName("network")

	subnetRes := &network.RDataSubnet{
		Type:        "cloud",
		NetworkID:   networkRes.TFID() + ".id",
		NetworkZone: "eu-central",
		IPRange:     "10.0.1.0/24",
	}
	subnetRes.SetRName("network")

	bastionRes := &server.RData{
		Name:         name + "-bastion",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{sshKeyRes.TFID() + ".id"},
		Networks: []server.RDataInlineNetwork{{
			NetworkID: networkRes.TFID() + ".id",
		}},
		PublicNet: map[string]interface{}{
			"ipv4_enabled": true,
			"ipv6_enabled": true,
		},
		UserData: `#cloud-config
users:
  - default
  - name: test
    shell: /bin/bash

runcmd:
  - echo "hello from bastion!"
`,
		DependsOn: []string{subnetRes.TFID()},
	}
	bastionRes.SetRName("bastion")

	hostRes := &server.RData{
		Name:         name + "-host",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{sshKeyRes.TFID() + ".id"},
		Networks: []server.RDataInlineNetwork{{
			NetworkID: networkRes.TFID() + ".id",
		}},
		PublicNet: map[string]interface{}{
			"ipv4_enabled": false,
			"ipv6_enabled": false,
		},
		UserData: `#cloud-config
users:
  - default
  - name: test
    shell: /bin/bash

runcmd:
  - echo "hello from host!"
`,
		DependsOn: []string{subnetRes.TFID()},
	}
	hostRes.SetRName("host")

	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sshKeyRes,
					"testdata/r/hcloud_network", networkRes,
					"testdata/r/hcloud_network_subnet", subnetRes,
					"testdata/r/hcloud_server", bastionRes,
					"testdata/r/hcloud_server", hostRes,
					"testdata/r/any",
					fmt.Sprintf(`
resource "terraform_data" "wait" {
  triggers_replace = [
    hcloud_server.bastion.id,
    hcloud_server.host.id,
  ]

  connection {
    type        = "ssh"
    user        = "root"
    host        = one(hcloud_server.host.network[*].ip)
    private_key = %q

    bastion_user = "root"
    bastion_host = hcloud_server.bastion.ipv4_address
  }

  provisioner "remote-exec" {
    inline = ["cloud-init status --wait --long || test $? -eq 2"]
  }
}
`, sshKeyRes.PrivateKey),
				),
			},
		},
	})
}

func TestAccServerResource_Firewalls(t *testing.T) {
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
		Type:        teste2e.TestServerType,
		Image:       teste2e.TestImage,
		FirewallIDs: []string{fw.TFID() + ".id"},
	}
	res.SetRName("server-firewall")
	res2 := &server.RData{
		Name:        "server-firewall",
		Type:        teste2e.TestServerType,
		Image:       teste2e.TestImage,
		FirewallIDs: []string{fw2.TFID() + ".id"},
	}
	res2.SetRName(res.RName())
	tmplMan := testtemplate.Manager{}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &s)),
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

func TestAccServerResource_PlacementGroup(t *testing.T) {
	var (
		pg  hcloud.PlacementGroup
		srv hcloud.Server
	)

	pgRes := placementgroup.NewRData(t, "server-test", "spread")

	srvRes := &server.RData{
		Name:             "server-placement-group",
		Type:             teste2e.TestServerType,
		Image:            teste2e.TestImage,
		PlacementGroupID: pgRes.TFID() + ".id",
	}
	srvRes.SetRName("server-placement-group")

	srvResNoPG := &server.RData{
		Name:  srvRes.Name,
		Type:  srvRes.Type,
		Image: srvRes.Image,
	}
	srvResNoPG.SetRName("server-placement-group")

	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
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
					resource.TestCheckResourceAttr(srvRes.TFID(), "name", fmt.Sprintf("server-placement-group--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(srvRes.TFID(), "server_type", srvRes.Type),
					resource.TestCheckResourceAttr(srvRes.TFID(), "image", srvRes.Image),
					testsupport.CheckResourceAttrFunc(srvRes.TFID(), "placement_group_id", func() string {
						return util.FormatID(pg.ID)
					}),
				),
			},
			{
				// Try to remove PG of running server -> error
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", pgRes,
					"testdata/r/hcloud_server", srvResNoPG,
				),
				ExpectError: regexp.MustCompile("removing a running server from a placement group is currently not supported in the provider.*"),
			},
			{
				// Remove Placement Group
				PreConfig: func() {
					ctx := context.TODO()
					// Removing PG is not support only in TF, we need to shut down the server manually beforehand
					client, err := testsupport.CreateClient()
					if err != nil {
						t.Errorf("PreConfig: failed to create client: %v", err)
						return
					}
					action, _, err := client.Server.Poweroff(ctx, &srv)
					if err != nil {
						t.Errorf("PreConfig: failed to power off server: %v", err)
						return
					}
					err = hcloudutil.WaitForAction(ctx, &client.Action, action)
					if err != nil {
						t.Errorf("PreConfig: power off server action failed: %v", err)
						return
					}
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", pgRes,
					"testdata/r/hcloud_server", srvResNoPG,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(srvResNoPG.TFID(), "status", "off"),
					resource.TestCheckResourceAttr(srvResNoPG.TFID(), "placement_group_id", "0"),
				),
			},
			{
				// Add Placement Group back
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", pgRes,
					"testdata/r/hcloud_server", srvRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(srvResNoPG.TFID(), "status", "running"),
					testsupport.CheckResourceAttrFunc(srvRes.TFID(), "placement_group_id", func() string {
						return util.FormatID(pg.ID)
					}),
				),
			},
		},
	})
}

func TestAccServerResource_Protection(t *testing.T) {
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
		Type:              teste2e.TestServerType,
		Image:             teste2e.TestImage,
		DeleteProtection:  true,
		RebuildProtection: true,
	}
	srvRes.SetRName("server-protection")

	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
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

func TestAccServerResource_EmptySSHKey(t *testing.T) {
	// Regression Test for https://github.com/hetznercloud/terraform-provider-hcloud/issues/727
	var srv hcloud.Server

	srvRes := &server.RData{
		Name:    "server-empty-ssh-key",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{"\"\""},
	}
	srvRes.SetRName("server-empty-ssh-key")

	tmplMan := testtemplate.Manager{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, &srv)),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", srvRes,
				),
				ExpectError: regexp.MustCompile("Invalid ssh key passed"),
			},
		},
	})
}

func isRecreated(newServer, oldServer *hcloud.Server) func() error {
	return func() error {
		if newServer.ID == oldServer.ID {
			return fmt.Errorf("new server is the same as server cert %d", oldServer.ID)
		}
		return nil
	}
}

func userDataHashSum(userData string) string {
	sum := sha1.Sum([]byte(userData)) // nolint: gosec
	return base64.StdEncoding.EncodeToString(sum[:])
}

func TestToPublicNetField(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		got, err := server.ToPublicNetField[int](map[string]any{"key": int(1)}, "key")
		require.NoError(t, err)
		require.Equal(t, int(1), got)
	})

	t.Run("bool", func(t *testing.T) {
		got, err := server.ToPublicNetField[bool](map[string]any{"key": true}, "key")
		require.NoError(t, err)
		require.Equal(t, true, got)
	})

	t.Run("int not found", func(t *testing.T) {
		got, err := server.ToPublicNetField[int](map[string]any{}, "key")
		require.EqualError(t, err, "ToPublicNetField: field does not contain key: key")
		require.Equal(t, int(0), got)
	})

	t.Run("bool not found", func(t *testing.T) {
		got, err := server.ToPublicNetField[bool](map[string]any{}, "key")
		require.EqualError(t, err, "ToPublicNetField: field does not contain key: key")
		require.Equal(t, false, got)
	})
}
