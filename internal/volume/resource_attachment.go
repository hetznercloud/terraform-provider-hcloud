package volume

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// AttachmentResourceType is the type name of the Hetzner Cloud Volume
// attachment resource.
const AttachmentResourceType = "hcloud_volume_attachment"

// AttachmentResource creates a Terraform schema for the
// hcloud_volume_attachmetn resource.
func AttachmentResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVolumeAttachmentCreate,
		ReadContext:   resourceVolumeAttachmentRead,
		DeleteContext: resourceVolumeAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"volume_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"automount": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceVolumeAttachmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var a *hcloud.Action

	c := m.(*hcloud.Client)

	volumeID := d.Get("volume_id")
	volume := &hcloud.Volume{ID: volumeID.(int)}

	serverID := d.Get("server_id")

	server := &hcloud.Server{ID: serverID.(int)}

	opts := hcloud.VolumeAttachOpts{
		Server: server,
	}
	if automount, ok := d.GetOk("automount"); ok {
		opts.Automount = hcloud.Bool(automount.(bool))
	}

	err := control.Retry(control.DefaultRetries, func() error {
		var err error

		a, _, err = c.Volume.AttachWithOpts(ctx, volume, opts)
		if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			return err
		}
		return control.AbortRetry(err)
	})
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	// Since a volume can only be attached to one server
	// we can use the volume id as volume attachment id.
	d.SetId(strconv.Itoa(volume.ID))
	return resourceVolumeAttachmentRead(ctx, d, m)
}

func resourceVolumeAttachmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	volumeID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Volume ID (%s) not found, removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	// 'volume_id' and 'server_id' is 'Required' and 'TypeInt'
	// therefore the cast should always work
	volume, _, err := client.Volume.GetByID(ctx, volumeID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if volume == nil {
		log.Printf("[WARN] Volume ID (%v) not found, removing volume attachment from state", d.Get("volume_id"))
		d.SetId("")
		return nil
	}
	// check if volume is attached to any server
	if volume.Server == nil {
		log.Printf("[WARN] Volume (%v) is not attached to a server, removing volume attachment from state", d.Get("volume_id"))
		d.SetId("")
		return nil
	}

	// when importing the resource the server_id is not given
	// because only the terraform ID (volume ID in this case)
	// is available, so we need to get the ID from the volume
	// instead of from the configuration
	serverID := d.Get("server_id").(int)
	if serverID == 0 {
		serverID = volume.Server.ID
	}

	server, _, err := client.Server.GetByID(ctx, serverID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil {
		log.Printf("[WARN] Server ID (%v) not found, removing volume attachment from state", d.Get("server_id"))
		d.SetId("")
		return nil
	}

	d.Set("server_id", volume.Server.ID)
	d.Set("volume_id", volume.ID)
	return nil
}

func resourceVolumeAttachmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	volumeID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	volume, _, err := c.Volume.GetByID(ctx, volumeID)
	if err != nil {
		log.Printf("[WARN] Volume ID (%v) not found, removing volume attachment from state", d.Get("volume_id"))
		d.SetId("")
		return nil
	}
	if volume == nil {
		log.Printf("[WARN] Volume ID (%v) not found, removing volume attachment from state", d.Get("volume_id"))
		d.SetId("")
		return nil
	}
	if volume.Server != nil {
		var a *hcloud.Action

		err := control.Retry(control.DefaultRetries, func() error {
			var err error

			a, _, err = c.Volume.Detach(ctx, volume)
			if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
				return err
			}
			return control.AbortRetry(err)
		})
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}

		if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	return nil
}
