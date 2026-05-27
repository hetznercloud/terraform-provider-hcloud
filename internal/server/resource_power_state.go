package server

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// PowerStateResourceType is the type name of the Hetzner Cloud Server Power State resource.
const PowerStateResourceType = "hcloud_server_power_state"

// PowerStateResource creates a Terraform schema for the hcloud_server_power_state resource.
func PowerStateResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerPowerStateCreate,
		ReadContext:   resourceServerPowerStateRead,
		UpdateContext: resourceServerPowerStateUpdate,
		DeleteContext: resourceServerPowerStateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the server whose power state should be managed.",
			},
			"state": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Desired power state of the server. Valid values are \"running\" and \"off\".",
				ValidateFunc: validation.StringInSlice([]string{
					serverPowerStateRunning,
					serverPowerStateOff,
				}, false),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Raw status returned by the Hetzner Cloud API.",
			},
		},
	}
}

func resourceServerPowerStateCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*hcloud.Client)
	serverID := util.CastInt64(d.Get("server_id"))
	d.SetId(util.FormatID(serverID))

	if err := setPowerState(ctx, client, &hcloud.Server{ID: serverID}, d.Get("state").(string)); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	return resourceServerPowerStateRead(ctx, d, m)
}

func resourceServerPowerStateRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*hcloud.Client)

	serverID, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid server power state id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	server, _, err := client.Server.GetByID(ctx, serverID)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			log.Printf("[WARN] Server (%s) not found, removing power state resource from state", d.Id())
			d.SetId("")
			return nil
		}
		return hcloudutil.ErrorToDiag(err)
	}
	if server == nil {
		d.SetId("")
		return nil
	}

	util.SetSchemaFromAttributes(d, map[string]any{
		"server_id": server.ID,
		"state":     powerStateFromServerStatus(server.Status),
		"status":    server.Status,
	})

	return nil
}

func resourceServerPowerStateUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*hcloud.Client)

	serverID, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid server power state id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	if d.HasChange("state") {
		if err := setPowerState(ctx, client, &hcloud.Server{ID: serverID}, d.Get("state").(string)); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	return resourceServerPowerStateRead(ctx, d, m)
}

func resourceServerPowerStateDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")
	return nil
}
