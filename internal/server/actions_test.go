package server_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/stretchr/testify/assert"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/server"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/sshkey"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/teste2e"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testsupport"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccServerActions(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	s := &hcloud.Server{}

	sk := sshkey.NewRData(t, "server-actions")

	res := &server.RData{
		Name:         "server-actions",
		Type:         teste2e.TestServerType,
		Image:        teste2e.TestImage,
		LocationName: teste2e.TestLocationName,
		SSHKeys:      []string{sk.TFID() + ".id"},
	}
	res.SetRName("default")

	resActionPoweroff := &server.AData{
		Type:     "poweroff",
		ServerID: res.TFID() + ".id",
	}
	resActionPoweroff.SetRName("default")

	resActionPoweron := testtemplate.DeepCopy(t, resActionPoweroff)
	resActionPoweron.Type = "poweron"

	resActionReboot := testtemplate.DeepCopy(t, resActionPoweroff)
	resActionReboot.Type = "reboot"

	resActionReset := testtemplate.DeepCopy(t, resActionPoweroff)
	resActionReset.Type = "reset"

	res.Raw = fmt.Sprintf(`
		lifecycle {
			action_trigger {
				events  = [after_create]
				actions = [
					%s,
					%s,
					%s,
					%s
				]
			}
		}
	`, resActionPoweroff.TFID(), resActionPoweron.TFID(), resActionReboot.TFID(), resActionReset.TFID())

	resource.ParallelTest(t, resource.TestCase{
		// Actions are only available in 1.14 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 teste2e.PreCheck(t),
		ProtoV6ProviderFactories: teste2e.ProtoV6ProviderFactories(),

		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_ssh_key", sk,
					"testdata/r/hcloud_server", res,
					"testdata/a/hcloud_server", resActionPoweroff,
					"testdata/a/hcloud_server", resActionPoweron,
					"testdata/a/hcloud_server", resActionReboot,
					"testdata/a/hcloud_server", resActionReset,
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckAPIResourcePresent(res.TFID(), testsupport.CopyAPIResource(s, server.GetAPIResource())),
					func(_ *terraform.State) error {
						client, err := testsupport.CreateClient()
						if err != nil {
							return err
						}

						actions, err := client.Server.Action.AllFor(context.Background(), s, hcloud.ActionListOpts{})
						if err != nil {
							return err
						}

						actionWithCommand := func(command string) func(*hcloud.Action) bool {
							return func(action *hcloud.Action) bool {
								return action.Command == command
							}
						}

						assert.True(t, slices.ContainsFunc(actions, actionWithCommand("stop_server")))
						assert.True(t, slices.ContainsFunc(actions, actionWithCommand("start_server")))
						assert.True(t, slices.ContainsFunc(actions, actionWithCommand("reboot_server")))
						assert.True(t, slices.ContainsFunc(actions, actionWithCommand("reset_server")))

						return nil
					},
				),
			},
		},
	})
}
