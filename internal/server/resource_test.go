package server_test

import (
	"crypto/sha1" // nolint: gosec
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testmux"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
)

func TestAccServerResource(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcServer hcloud.Server

	resSSHKey := sshkey.NewRData(t, "server-basic")

	res1 := &server.RData{
		Name:                   "server-basic",
		Type:                   teste2e.TestServerType,
		Image:                  teste2e.TestImage,
		SSHKeys:                []string{resSSHKey.TFID() + ".id"},
		ShutdownBeforeDeletion: true,
	}
	res1.SetRName("server-basic")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.Name += "-renamed"
	res2.Labels = map[string]string{"foo": "bar"}
	res2.Backups = true
	res2.ShutdownBeforeDeletion = false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-basic--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
					resource.TestCheckResourceAttrSet(res1.TFID(), "location"),
					resource.TestCheckResourceAttrSet(res1.TFID(), "datacenter"),
					resource.TestCheckResourceAttrPair(resSSHKey.TFID(), "id", res1.TFID(), "ssh_keys.0"),
					resource.TestCheckResourceAttrSet(res1.TFID(), "ipv4_address"),
					resource.TestCheckResourceAttrSet(res1.TFID(), "ipv6_address"),
					resource.TestCheckResourceAttrSet(res1.TFID(), "ipv6_network"),
					resource.TestCheckResourceAttr(res1.TFID(), "status", string(hcloud.ServerStatusRunning)),
					resource.TestCheckResourceAttrSet(res1.TFID(), "primary_disk_size"),
					resource.TestCheckResourceAttr(res1.TFID(), "placement_group_id", "0"),
				),
			},
			{
				// Try to import the newly created Server
				ResourceName:      res1.TFID(),
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
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res2.TFID(), "name", fmt.Sprintf("server-basic-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res2.TFID(), "server_type", res2.Type),
					resource.TestCheckResourceAttr(res2.TFID(), "image", res2.Image),
					resource.TestCheckResourceAttr(res2.TFID(), "labels.foo", "bar"),
					resource.TestCheckResourceAttr(res2.TFID(), "backups", "true"),
				),
			},
			{
				// Revert the server to the original state to test the other direction for various options
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res1,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res1.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res2.TFID(), "name", fmt.Sprintf("server-basic--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res2.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res2.TFID(), "image", res1.Image),
					resource.TestCheckNoResourceAttr(res2.TFID(), "labels.foo"),
					resource.TestCheckResourceAttr(res2.TFID(), "backups", "false"),
				),
			},
		},
	})
}

func TestAccServerResource_UnavailableServerType(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	res := &server.RData{
		Name:  "unavailable-server-type",
		Type:  "1",
		Image: teste2e.TestImage,
	}
	res.SetRName(res.Name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res,
				),
				ExpectError: regexp.MustCompile(`Server Type "cx11" is unavailable in all locations and can no longer be ordered`),
			},
		},
	})
}

func TestAccServerResource_ImageID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hsServer hcloud.Server

	resSSHKey := sshkey.NewRData(t, "server-image-id")

	resImage := &image.DData{
		ImageName:    teste2e.TestImage,
		Architecture: hcloud.ArchitectureX86,
	}
	resImage.SetRName("server-image-id")

	res := &server.RData{
		Name:    "server-image-id",
		Type:    teste2e.TestServerType,
		Image:   fmt.Sprintf("${%s.id}", resImage.TFID()),
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	res.SetRName("server-image-id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/d/hcloud_image", resImage,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hsServer)),
				),
			},
		},
	})
}

func TestAccServerResource_Resize(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcServer hcloud.Server

	resSSHKey := sshkey.NewRData(t, "server-resize")

	res1 := &server.RData{
		Name:    "server-resize",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	res1.SetRName("server-resize")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.Type = teste2e.TestServerTypeUpgrade
	res2.KeepDisk = true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
				),
			},
			{
				// Update the Server created in the previous step by
				// setting all optional fields.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res2.TFID(), "name", fmt.Sprintf("server-resize--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res2.TFID(), "server_type", res2.Type),
					resource.TestCheckResourceAttr(res2.TFID(), "image", res2.Image),
				),
			},
		},
	})
}

func TestAccServerResource_ChangeUserData(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcServer1, hcServer2 hcloud.Server

	sshKeyRes := sshkey.NewRData(t, "server-userdata")

	res1 := &server.RData{
		Name:     "server-userdata",
		Type:     teste2e.TestServerType,
		Image:    teste2e.TestImage,
		SSHKeys:  []string{sshKeyRes.TFID() + ".id"},
		UserData: "stuff",
	}
	res1.SetRName("server-userdata")

	// Update user data to force a replacement
	res2 := testtemplate.DeepCopy(t, res1)
	res2.UserData = "updated stuff"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sshKeyRes,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer1)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
					resource.TestCheckResourceAttr(res1.TFID(), "user_data", userDataHashSum(res1.UserData+"\n")),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sshKeyRes,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testsupport.CheckResourceExists(res2.TFID(), server.ByID(t, &hcServer2)),
					resource.TestCheckResourceAttr(res2.TFID(), "name", fmt.Sprintf("server-userdata--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res2.TFID(), "server_type", res2.Type),
					resource.TestCheckResourceAttr(res2.TFID(), "image", res2.Image),
					resource.TestCheckResourceAttr(res2.TFID(), "user_data", userDataHashSum(res2.UserData+"\n")),
				),
			},
		},
	})
}

func TestAccServerResource_ISO(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcServer hcloud.Server

	resSSHKey := sshkey.NewRData(t, "server-iso")

	res1 := &server.RData{
		Name:     "server-iso",
		Type:     teste2e.TestServerType,
		Image:    teste2e.TestImage,
		UserData: "stuff",
		ISO:      "8637", // Windows Server 2022 English
		SSHKeys:  []string{resSSHKey.TFID() + ".id"},
	}
	res1.SetRName("server-iso")

	// Update ISO
	res2 := testtemplate.DeepCopy(t, res1)
	res2.ISO = "8638" // Windows Server 2022 German

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-iso--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
					resource.TestCheckResourceAttr(res1.TFID(), "iso", res1.ISO),
				),
			},
			{
				// Update ISO
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res1.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "iso", res2.ISO),
				),
			},
		},
	})
}

func TestAccServerResource_Rescue(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcServer hcloud.Server

	resSSHKey := sshkey.NewRData(t, "server-rescue")

	res := &server.RData{
		Name:    "server-rescue",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{resSSHKey.TFID() + ".id"},
	}
	res.SetRName("server-rescue")

	resRescue := testtemplate.DeepCopy(t, res)
	resRescue.Rescue = string(hcloud.ServerRescueTypeLinux64)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server with rescue.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resRescue,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resRescue.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(resRescue.TFID(), "rescue", resRescue.Rescue),
				),
			},
			{
				// Disable server rescue.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res.TFID(), "rescue", ""),
				),
			},
			{
				// Enable rescue on existing server.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", resRescue,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resRescue.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res.TFID(), "rescue", resRescue.Rescue),
				),
			},
		},
	})
}

func TestAccServerResource_DirectAttachToNetwork(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetworkA hcloud.Network
		hcNetworkB hcloud.Network
		hcServer   hcloud.Server
	)

	resSSHKey := sshkey.NewRData(t, "server-direct-attach-network")

	nws := network.NewBlueprint(t)

	// Resource from which copies for every steps are made.
	res := &server.RData{
		Name:         "server-direct-attach",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
	}
	res.SetRName(res.Name)

	// Create a new server and directly attach it to a network.
	res1 := testtemplate.DeepCopy(t, res)
	res1.Networks = []server.RDataInlineNetwork{{
		NetworkID: nws.NetworkA.TFID() + ".id",
		IP:        "10.0.1.5",
		AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
	}}
	res1.DependsOn = []string{nws.SubnetA1.TFID()}

	// Fail when using the same network twice
	res2 := testtemplate.DeepCopy(t, res)
	res2.Networks = []server.RDataInlineNetwork{
		{
			NetworkID: nws.NetworkA.TFID() + ".id",
			IP:        "10.0.1.5",
			AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
		},
		{
			NetworkID: nws.NetworkA.TFID() + ".id",
			IP:        "10.0.1.8",
		},
	}
	res2.DependsOn = []string{nws.SubnetA1.TFID()}

	// Change the IP of the server
	res3 := testtemplate.DeepCopy(t, res)
	res3.Networks = []server.RDataInlineNetwork{{
		NetworkID: nws.NetworkA.TFID() + ".id",
		IP:        "10.0.1.4",
		AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
	}}
	res3.DependsOn = []string{nws.SubnetA1.TFID()}

	// Change the Alias IPs of the server
	res4 := testtemplate.DeepCopy(t, res)
	res4.Networks = []server.RDataInlineNetwork{{
		NetworkID: nws.NetworkA.TFID() + ".id",
		IP:        "10.0.1.4",
		AliasIPs:  []string{"10.0.1.5", "10.0.1.7"},
	}}
	res4.DependsOn = []string{nws.SubnetA1.TFID()}

	// Detach the server from the network.
	res5 := testtemplate.DeepCopy(t, res)

	// Remove all networks
	res6 := testtemplate.DeepCopy(t, res)

	// Attach to two new networks at the same time
	res7 := testtemplate.DeepCopy(t, res)
	res7.Networks = []server.RDataInlineNetwork{
		{
			NetworkID: nws.NetworkA.TFID() + ".id",
			IP:        "10.0.1.5",
			AliasIPs:  []string{"10.0.1.6", "10.0.1.7"},
		},
		{
			NetworkID: nws.NetworkB.TFID() + ".id",
			IP:        "172.16.1.5",
			AliasIPs:  []string{"172.16.1.6", "172.16.1.7"},
		},
	}
	res7.DependsOn = []string{nws.SubnetA1.TFID(), nws.SubnetB1.TFID()}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetworkA)),
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkA, "10.0.1.5", "10.0.1.6", "10.0.1.7")),
					resource.TestCheckResourceAttr(res.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res.TFID(), "network.0.ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(res.TFID(), "network.0.alias_ips.#", "2"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res2,
				),
				ExpectError: regexp.MustCompile(`server is only allowed to be attached to each network once: \d+`),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res3,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetworkA)),
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkA, "10.0.1.4", "10.0.1.6", "10.0.1.7")),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res4,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetworkA)),
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkA, "10.0.1.4", "10.0.1.5", "10.0.1.7")),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res5,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(func() error {
						assert.Empty(t, hcServer.PrivateNet)
						return nil
					}),
				),
			},
			{
				// Remove networks for next step
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_server", res6,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(func() error {
						assert.Empty(t, hcServer.PrivateNet)
						return nil
					}),
				),
			},
			{
				// Attach to two new networks at the same time
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network", nws.NetworkB,
					"testdata/r/hcloud_network_subnet", nws.SubnetB1,
					"testdata/r/hcloud_server", res7,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetworkA)),
					testsupport.CheckResourceExists(nws.NetworkB.TFID(), network.ByID(t, &hcNetworkB)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkA, "10.0.1.5", "10.0.1.6", "10.0.1.7")),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkB, "172.16.1.5", "172.16.1.6", "172.16.1.7")),
				),
			},
		},
	})
}

func TestAccServerResource_DirectAttachToNetworkSubnetID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork hcloud.Network
		hcServer  hcloud.Server
	)

	nws := network.NewBlueprint(t)

	resSSHKey := sshkey.NewRData(t, "server-direct-attach-subnet")

	res := &server.RData{
		Name:         "server-subnet",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},

		DependsOn: []string{nws.SubnetA1.TFID(), nws.SubnetA2.TFID()},
	}
	res.SetRName(res.Name)

	// Step 1: Server with subnet_id
	res1 := testtemplate.DeepCopy(t, res)
	res1.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA1.TFID() + ".id",
	}}

	// Step 2: Server with both network_id and subnet_id
	res2 := testtemplate.DeepCopy(t, res)
	res2.Networks = []server.RDataInlineNetwork{{
		NetworkID: nws.NetworkA.TFID() + ".id",
		SubnetID:  nws.SubnetA1.TFID() + ".id",
	}}

	// Step 3: Fail when network_id and subnet_id mismatch
	res3 := testtemplate.DeepCopy(t, res)
	res3.Networks = []server.RDataInlineNetwork{{
		NetworkID: "12345",
		SubnetID:  nws.SubnetA1.TFID() + ".id",
	}}

	// Step 4: Fail when trying to attach to the same network twice using subnet_id
	res4 := testtemplate.DeepCopy(t, res)
	res4.Networks = []server.RDataInlineNetwork{
		{
			SubnetID: nws.SubnetA1.TFID() + ".id",
		},
		{
			SubnetID: nws.SubnetA2.TFID() + ".id",
			IP:       "10.0.2.1",
		},
	}

	// Step 5: Fail when IP is outside subnet range
	res5 := testtemplate.DeepCopy(t, res)
	res5.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA2.TFID() + ".id",
		IP:       "10.0.1.1", // Outside 10.0.2.0/24 range
	}}

	// Step 6: Fail when neither network_id nor subnet_id is specified
	res6 := testtemplate.DeepCopy(t, res)
	res6.Networks = []server.RDataInlineNetwork{{
		IP: "10.0.1.1",
	}}

	// Step 7: Server with subnet_id and explicit IP
	res7 := testtemplate.DeepCopy(t, res)
	res7.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA1.TFID() + ".id",
		IP:       "10.0.1.10",
	}}

	// Step 8: Server with new subnet_id and explicit IP
	res8 := testtemplate.DeepCopy(t, res)
	res8.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA2.TFID() + ".id",
		IP:       "10.0.2.20",
	}}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.1")),
					resource.TestCheckResourceAttr(res1.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res1.TFID(), "network.0.ip", "10.0.1.1"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res2,
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res3,
				),
				ExpectError: regexp.MustCompile(`subnet_id \(\d+-10\.0\.1\.0\/24\) does not belong to the specified network_id \(12345\)`),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res4,
				),
				ExpectError: regexp.MustCompile(`server is only allowed to be attached to each network once: \d+`),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res5,
				),
				ExpectError: regexp.MustCompile(`server IP \(10\.0\.1\.1\) is outside subnet IP range \(10\.0\.2\.0/24\)`),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res6,
				),
				ExpectError: regexp.MustCompile(`must specify either network_id or subnet_id`),
			},
			{
				// Update server to attach with specific IP in the same subnet
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res7,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.10")),
					resource.TestCheckResourceAttr(res1.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res1.TFID(), "network.0.ip", "10.0.1.10"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res8,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(res8.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.2.20")),
					resource.TestCheckResourceAttr(res8.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res8.TFID(), "network.0.ip", "10.0.2.20"),
				),
			},
		},
	})
}

func TestAccServerResource_DirectAttachToNetworkID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork hcloud.Network
		hcServer  hcloud.Server
	)

	nws := network.NewBlueprint(t)

	resSSHKey := sshkey.NewRData(t, "server-network-id-compat")

	// Server using network_id only
	res := &server.RData{
		Name:         "server-network-id",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
		Networks: []server.RDataInlineNetwork{{
			NetworkID: nws.NetworkA.TFID() + ".id",
			IP:        "10.0.1.10",
		}},
		DependsOn: []string{nws.SubnetA1.TFID()},
	}
	res.SetRName("server-network-id")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create server using network_id only
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.SubnetA1.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.10")),
					resource.TestCheckResourceAttr(res.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res.TFID(), "network.0.ip", "10.0.1.10"),
					resource.TestCheckResourceAttrSet(res.TFID(), "network.0.network_id"),
				),
			},
		},
	})
}

func TestAccServerResource_PrimaryIPs(t *testing.T) {
	// This test focus on the server primary ips.

	tmplMan := testtemplate.Manager{}

	var (
		hcServer    hcloud.Server
		hcPrimaryIP hcloud.PrimaryIP
	)

	ips := primaryip.NewBlueprint(t)
	nws := network.NewBlueprint(t)

	resSSHKey := sshkey.NewRData(t, "server")

	// Create server with unmanaged primary ips and network
	res1 := &server.RData{
		Name:         "primary-ips-test",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
		Networks: []server.RDataInlineNetwork{{
			NetworkID: nws.NetworkA.TFID() + ".id",
			IP:        "10.0.1.5",
		}},
		PublicNet: map[string]any{
			"ipv4_enabled": true,
			"ipv6_enabled": true,
		},
		DependsOn: []string{nws.SubnetA1.TFID()},
	}
	res1.SetRName(res1.Name)

	// - Remove unmanaged primary ipv4
	// - Remove unmanaged primary ipv6
	res2 := testtemplate.DeepCopy(t, res1)
	res2.PublicNet = map[string]any{
		"ipv4_enabled": false,
		"ipv6_enabled": false,
	}

	// - Add managed primary ipv4
	res3 := testtemplate.DeepCopy(t, res2)
	res3.PublicNet = map[string]any{
		"ipv4_enabled": true,
		"ipv4":         ips.PrimaryIPv4A.TFID() + ".id",
		"ipv6_enabled": false,
	}

	// - Add unmanaged primary ipv6
	res4 := testtemplate.DeepCopy(t, res3)
	res4.PublicNet = map[string]any{
		"ipv4_enabled": true,
		"ipv4":         ips.PrimaryIPv4A.TFID() + ".id",
		"ipv6_enabled": true,
	}

	// Remove public net config, and rely on the defaults:
	// - Remove managed primary ipv4
	// - Add unmanaged primary ipv4
	res5 := testtemplate.DeepCopy(t, res4)
	res5.PublicNet = nil

	// - Remove unmanaged primary ipv4
	// - Remove unmanaged primary ipv6
	// - Add managed primary ipv6
	res6 := testtemplate.DeepCopy(t, res5)
	res6.PublicNet = map[string]any{
		"ipv4_enabled": false,
		"ipv6_enabled": true,
		"ipv6":         ips.PrimaryIPv6C.TFID() + ".id",
	}

	// - Remove managed primary ipv6
	// - Add unmanaged primary ipv6
	res7 := testtemplate.DeepCopy(t, res6)
	res7.PublicNet = map[string]any{
		"ipv4_enabled": false,
		"ipv6_enabled": true,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "network.#", "1"),

					testsupport.LiftTCF(func() error {
						assert.NotEqual(t, int64(0), hcServer.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), hcServer.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res2.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(func() error {
						assert.Nil(t, hcServer.PublicNet.IPv4.IP)
						assert.Nil(t, hcServer.PublicNet.IPv6.IP)
						return nil
					}),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_primary_ip", ips.PrimaryIPv4A,
					"testdata/r/hcloud_server", res3,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res3.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res3.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(ips.PrimaryIPv4A.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
					testsupport.LiftTCF(func() error {
						assert.Equal(t, hcPrimaryIP.AssigneeID, hcServer.ID)
						assert.Equal(t, hcPrimaryIP.ID, hcServer.PublicNet.IPv4.ID)
						assert.Equal(t, int64(0), hcServer.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_primary_ip", ips.PrimaryIPv4A,
					"testdata/r/hcloud_server", res4,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res4.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res4.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(ips.PrimaryIPv4A.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
					testsupport.LiftTCF(func() error {
						assert.Equal(t, hcPrimaryIP.AssigneeID, hcServer.ID)
						assert.Equal(t, hcPrimaryIP.ID, hcServer.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), hcServer.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_primary_ip", ips.PrimaryIPv4A,
					"testdata/r/hcloud_server", res5,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res5.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res5.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(ips.PrimaryIPv4A.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
					testsupport.LiftTCF(func() error {
						assert.NotEqual(t, hcPrimaryIP.ID, hcServer.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), hcServer.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), hcServer.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_primary_ip", ips.PrimaryIPv6C,
					"testdata/r/hcloud_server", res6,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res6.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res6.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(ips.PrimaryIPv6C.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
					testsupport.LiftTCF(func() error {
						assert.Equal(t, hcPrimaryIP.ID, hcServer.PublicNet.IPv6.ID)
						assert.Equal(t, int64(0), hcServer.PublicNet.IPv4.ID)
						return nil
					}),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_primary_ip", ips.PrimaryIPv6C,
					"testdata/r/hcloud_server", res7,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// TODO: The resource should be updated.
						plancheck.ExpectResourceAction(res7.TFID(), plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res7.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(ips.PrimaryIPv6C.TFID(), primaryip.ByID(t, &hcPrimaryIP)),
					testsupport.LiftTCF(func() error {
						assert.NotEqual(t, hcPrimaryIP.ID, hcServer.PublicNet.IPv4.ID)
						// TODO: The managed primary ipv6 should have been removed and
						// replaced with an unmanaged one.
						assert.Equal(t, hcPrimaryIP.ID, hcServer.PublicNet.IPv6.ID)
						assert.Equal(t, int64(0), hcServer.PublicNet.IPv4.ID)
						assert.NotEqual(t, int64(0), hcServer.PublicNet.IPv6.ID)
						return nil
					}),
				),
			},
		},
	})
}

func TestAccServerResource_PrivateNetworkBastion(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	name := "server-private-network-bastion"

	resSSHKey := sshkey.NewRData(t, name)

	nws := network.NewBlueprint(t)

	bastionRes := &server.RData{
		Name:         name + "-bastion",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
		Networks: []server.RDataInlineNetwork{{
			NetworkID: nws.NetworkA.TFID() + ".id",
		}},
		PublicNet: map[string]any{
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
		DependsOn: []string{nws.SubnetA1.TFID()},
	}
	bastionRes.SetRName("bastion")

	hostRes := &server.RData{
		Name:         name + "-host",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
		Networks: []server.RDataInlineNetwork{{
			NetworkID: nws.NetworkA.TFID() + ".id",
		}},
		PublicNet: map[string]any{
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
		DependsOn: []string{nws.SubnetA1.TFID()},
	}
	hostRes.SetRName("host")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
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
`, resSSHKey.PrivateKey),
				),
			},
		},
	})
}

func TestAccServerResource_Firewalls(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var hcServer hcloud.Server

	resFirewallA := firewall.NewRData(t, "server-test-a", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "icmp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
		},
	}, nil)

	resFirewallB := firewall.NewRData(t, "server-test-b", []firewall.RDataRule{
		{
			Direction: "in",
			Protocol:  "tcp",
			SourceIPs: []string{"0.0.0.0/0", "::/0"},
			Port:      "1-65535",
		},
	}, nil)

	res1 := &server.RData{
		Name:        "server-firewall",
		Type:        teste2e.TestServerType,
		Image:       teste2e.TestImage,
		FirewallIDs: []string{resFirewallA.TFID() + ".id"},
	}
	res1.SetRName("server-firewall")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.FirewallIDs = []string{resFirewallB.TFID() + ".id"}

	res3 := testtemplate.DeepCopy(t, res2)
	res3.FirewallIDs = nil

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_firewall", resFirewallA,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
					resource.TestCheckResourceAttr(res1.TFID(), "firewall_ids.#", "1"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_firewall", resFirewallA,
					"testdata/r/hcloud_firewall", resFirewallB,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
					resource.TestCheckResourceAttr(res1.TFID(), "firewall_ids.#", "1"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_firewall", resFirewallA,
					"testdata/r/hcloud_firewall", resFirewallB,
					"testdata/r/hcloud_server", res3,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-firewall--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
					resource.TestCheckResourceAttr(res1.TFID(), "firewall_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccServerResource_PlacementGroup(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcPlacementGroup hcloud.PlacementGroup
		hcServer         hcloud.Server
	)

	resPlacementGroup := placementgroup.NewRData(t, "server-test", "spread")

	resWithoutPG := &server.RData{
		Name:  "server-placement-group",
		Type:  teste2e.TestServerType,
		Image: teste2e.TestImage,
	}
	resWithoutPG.SetRName("server-placement-group")

	resWithPG := testtemplate.DeepCopy(t, resWithoutPG)
	resWithPG.PlacementGroupID = resPlacementGroup.TFID() + ".id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", resPlacementGroup,
					"testdata/r/hcloud_server", resWithPG,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(resWithPG.TFID(), server.ByID(t, &hcServer)),
					testsupport.CheckResourceExists(resPlacementGroup.TFID(), placementgroup.ByID(t, &hcPlacementGroup)),
					resource.TestCheckResourceAttr(resWithPG.TFID(), "name", fmt.Sprintf("server-placement-group--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resWithPG.TFID(), "server_type", resWithPG.Type),
					resource.TestCheckResourceAttr(resWithPG.TFID(), "image", resWithPG.Image),
					testsupport.CheckResourceAttrFunc(resWithPG.TFID(), "placement_group_id", func() string {
						return util.FormatID(hcPlacementGroup.ID)
					}),
				),
			},
			{
				// Try to remove PG of running server -> error
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", resPlacementGroup,
					"testdata/r/hcloud_server", resWithoutPG,
				),
				ExpectError: regexp.MustCompile("removing a running server from a placement group is currently not supported in the provider.*"),
			},
			{
				// Remove Placement Group
				PreConfig: func() {
					ctx := t.Context()
					// Removing PG is not support only in TF, we need to shut down the server manually beforehand
					client, err := testsupport.CreateClient()
					if err != nil {
						t.Errorf("PreConfig: failed to create client: %v", err)
						return
					}
					action, _, err := client.Server.Poweroff(ctx, &hcServer)
					if err != nil {
						t.Errorf("PreConfig: failed to power off server: %v", err)
						return
					}
					err = client.Action.WaitFor(ctx, action)
					if err != nil {
						t.Errorf("PreConfig: power off server action failed: %v", err)
						return
					}
				},
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", resPlacementGroup,
					"testdata/r/hcloud_server", resWithoutPG,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resWithoutPG.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resWithoutPG.TFID(), "status", "off"),
					resource.TestCheckResourceAttr(resWithoutPG.TFID(), "placement_group_id", "0"),
				),
			},
			{
				// Add Placement Group back
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_placement_group", resPlacementGroup,
					"testdata/r/hcloud_server", resWithPG,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resWithPG.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resWithoutPG.TFID(), "status", "running"),
					testsupport.CheckResourceAttrFunc(resWithPG.TFID(), "placement_group_id", func() string {
						return util.FormatID(hcPlacementGroup.ID)
					}),
				),
			},
		},
	})
}

func TestAccServerResource_Protection(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcServer hcloud.Server
	)

	res1 := &server.RData{
		Name:              "server-protection",
		Type:              teste2e.TestServerType,
		Image:             teste2e.TestImage,
		DeleteProtection:  true,
		RebuildProtection: true,
	}
	res1.SetRName("server-protection")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.DeleteProtection = false
	res2.RebuildProtection = false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values
				// only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					resource.TestCheckResourceAttr(res1.TFID(), "name", fmt.Sprintf("server-protection--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res1.TFID(), "server_type", res1.Type),
					resource.TestCheckResourceAttr(res1.TFID(), "image", res1.Image),
					resource.TestCheckResourceAttr(res1.TFID(), "delete_protection", "true"),
					resource.TestCheckResourceAttr(res1.TFID(), "rebuild_protection", "true"),
				),
			},
			{
				// Update delete protection
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res1.TFID(), "delete_protection", "false"),
					resource.TestCheckResourceAttr(res1.TFID(), "rebuild_protection", "false"),
				),
			},
		},
	})
}

func TestAccServerResource_EmptySSHKey(t *testing.T) {
	// Regression test for https://github.com/hetznercloud/terraform-provider-hcloud/issues/727

	tmplMan := testtemplate.Manager{}

	res := &server.RData{
		Name:    "server-empty-ssh-key",
		Type:    teste2e.TestServerType,
		Image:   teste2e.TestImage,
		SSHKeys: []string{"\"\""},
	}
	res.SetRName("server-empty-ssh-key")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create a new Server using the required values only.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", res,
				),
				ExpectError: regexp.MustCompile("Invalid ssh key passed"),
			},
		},
	})
}

func TestAccServerResource_DatacenterToLocation(t *testing.T) {
	// Test for the "datacenter" deprecation, to make sure that its possible to move to "location" attribute
	// See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters

	tmplMan := testtemplate.Manager{}

	res1 := &server.RData{
		Name:       "server-dc-to-location",
		Type:       teste2e.TestServerType,
		Image:      teste2e.TestImage,
		Datacenter: teste2e.TestDataCenter,
	}
	res1.SetRName("dc_to_location")

	res2 := testtemplate.DeepCopy(t, res1)
	res2.Datacenter = ""
	res2.LocationName = teste2e.TestLocationName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create server in Datacenter.
				Config: tmplMan.Render(t, "testdata/r/hcloud_server", res1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(res1.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
			{
				// Change config to Location.
				Config: tmplMan.Render(t, "testdata/r/hcloud_server", res2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(res2.TFID(), tfjsonpath.New("datacenter"), knownvalue.StringExact(teste2e.TestDataCenter)),
					statecheck.ExpectKnownValue(res2.TFID(), tfjsonpath.New("location"), knownvalue.StringExact(teste2e.TestLocationName)),
				},
			},
		},
	})
}

func TestAccServerResource_MigrateNetworkIDToSubnetID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork hcloud.Network
		hcServer  hcloud.Server
	)

	nws := network.NewBlueprint(t)

	resSSHKey := sshkey.NewRData(t, "migrate-to-subnet")

	res := &server.RData{
		Name:         "migrate-to-subnet",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
	}
	res.SetRName(res.Name)

	// Step 1: Create using network_id
	res1 := testtemplate.DeepCopy(t, res)
	res1.Networks = []server.RDataInlineNetwork{{
		NetworkID: nws.NetworkA.TFID() + ".id",
		IP:        "10.0.1.5",
	}}
	res1.DependsOn = []string{nws.SubnetA1.TFID()}

	// Step 2: Switch to subnet_id, keep same IP
	res2 := testtemplate.DeepCopy(t, res)
	res2.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA1.TFID() + ".id",
		IP:       "10.0.1.5",
	}}

	// Step 3: Remove IP field, server should keep the same IP
	res3 := testtemplate.DeepCopy(t, res)
	res3.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA1.TFID() + ".id",
	}}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create with network_id
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res1.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res1.TFID(), "network.0.ip", "10.0.1.5"),
					resource.TestCheckResourceAttrSet(res1.TFID(), "network.0.network_id"),
				),
			},
			{
				// Migrate to subnet_id, server should update, not replace
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res2.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res2.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res2.TFID(), "network.0.ip", "10.0.1.5"),
				),
			},
			{
				// Remove IP field, server should keep the same IP (no re-attach)
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res3,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res3.TFID(), plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res3.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res3.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res3.TFID(), "network.0.ip", "10.0.1.5"),
				),
			},
			{
				// Rollback to network_id with same IP, should not trigger changes
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_server", res1,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res1.TFID(), plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res1.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res1.TFID(), "network.0.ip", "10.0.1.5"),
					resource.TestCheckResourceAttrSet(res1.TFID(), "network.0.network_id"),
				),
			},
		},
	})
}

func TestAccServerResource_SwitchSubnet(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetwork hcloud.Network
		hcServer  hcloud.Server
	)

	resSSHKey := sshkey.NewRData(t, "switch-subnet")

	nws := network.NewBlueprint(t)

	res := &server.RData{
		Name:         "switch-subnet",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
	}
	res.SetRName(res.Name)

	// Attach to subnet1
	res1 := testtemplate.DeepCopy(t, res)
	res1.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA1.TFID() + ".id",
		IP:       "10.0.1.5",
	}}

	// Switch to subnet2
	res2 := testtemplate.DeepCopy(t, res)
	res2.Networks = []server.RDataInlineNetwork{{
		SubnetID: nws.SubnetA2.TFID() + ".id",
		IP:       "10.0.2.5",
	}}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create with subnet1
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res1,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res1.TFID(), plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(res1.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.1.5")),
					resource.TestCheckResourceAttr(res1.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res1.TFID(), "network.0.ip", "10.0.1.5"),
				),
			},
			{
				// Switch to subnet2, IP changes so the server detaches and re-attaches
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network_subnet", nws.SubnetA2,
					"testdata/r/hcloud_server", res2,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(res2.TFID(), plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetwork)),
					testsupport.CheckResourceExists(res2.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetwork, "10.0.2.5")),
					resource.TestCheckResourceAttr(res2.TFID(), "network.#", "1"),
					resource.TestCheckResourceAttr(res2.TFID(), "network.0.ip", "10.0.2.5"),
				),
			},
		},
	})
}

func TestAccServerResource_MixedNetworkAndSubnetID(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	var (
		hcNetworkA hcloud.Network
		hcNetworkB hcloud.Network
		hcServer   hcloud.Server
	)

	nws := network.NewBlueprint(t)

	resSSHKey := sshkey.NewRData(t, "server-mixed-attach")

	res := &server.RData{
		Name:         "server-mixed-networks",
		Type:         teste2e.TestServerType,
		LocationName: teste2e.TestLocationName,
		Image:        teste2e.TestImage,
		SSHKeys:      []string{resSSHKey.TFID() + ".id"},
	}
	res.SetRName("server-mixed-networks")

	// Server: network1 via network_id, network2 via subnet_id
	res1 := testtemplate.DeepCopy(t, res)
	res1.Networks = []server.RDataInlineNetwork{
		{
			NetworkID: nws.NetworkA.TFID() + ".id",
			IP:        "10.0.1.5",
		},
		{
			SubnetID: nws.SubnetB1.TFID() + ".id",
			IP:       "172.16.1.5",
		},
	}
	res1.DependsOn = []string{nws.SubnetA1.TFID(), nws.SubnetB1.TFID()}

	// Detach network1, keep network2 via subnet_id
	res2 := testtemplate.DeepCopy(t, res)
	res2.Networks = []server.RDataInlineNetwork{
		{
			SubnetID: nws.SubnetB1.TFID() + ".id",
			IP:       "172.16.1.5",
		},
	}
	res2.DependsOn = []string{nws.SubnetA1.TFID(), nws.SubnetB1.TFID()}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: testmux.ProtoV6ProviderFactories(),
		CheckDestroy:             testsupport.CheckAPIResourceAllAbsent(server.ResourceType, server.GetAPIResource()),
		Steps: []resource.TestStep{
			{
				// Create with mixed attachment types
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network", nws.NetworkB,
					"testdata/r/hcloud_network_subnet", nws.SubnetB1,
					"testdata/r/hcloud_server", res1,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(nws.NetworkA.TFID(), network.ByID(t, &hcNetworkA)),
					testsupport.CheckResourceExists(nws.NetworkB.TFID(), network.ByID(t, &hcNetworkB)),
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkA, "10.0.1.5")),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkB, "172.16.1.5")),
					resource.TestCheckResourceAttr(res.TFID(), "network.#", "2"),
				),
			},
			{
				// Detach from network1, keep network2 via subnet_id
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", resSSHKey,
					"testdata/r/hcloud_network", nws.NetworkA,
					"testdata/r/hcloud_network_subnet", nws.SubnetA1,
					"testdata/r/hcloud_network", nws.NetworkB,
					"testdata/r/hcloud_network_subnet", nws.SubnetB1,
					"testdata/r/hcloud_server", res2,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.TFID(), server.ByID(t, &hcServer)),
					testsupport.LiftTCF(hasServerNetwork(t, &hcServer, &hcNetworkB, "172.16.1.5")),
					testsupport.LiftTCF(func() error {
						assert.Nil(t, hcServer.PrivateNetFor(&hcNetworkA))
						return nil
					}),
					resource.TestCheckResourceAttr(res.TFID(), "network.#", "1"),
				),
			},
		},
	})
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
