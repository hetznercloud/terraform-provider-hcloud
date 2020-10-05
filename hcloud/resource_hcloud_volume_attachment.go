package hcloud

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceVolumeAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceVolumeAttachmentCreate,
		Read:   resourceVolumeAttachmentRead,
		Delete: resourceVolumeAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceVolumeAttachmentCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

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

	action, _, err := client.Volume.AttachWithOpts(ctx, volume, opts)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			log.Printf("[INFO] Server (%v) locked, retrying in one second", serverID)
			time.Sleep(time.Second)
			return resourceVolumeAttachmentCreate(d, m)
		}
		return err
	}
	if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
		return err
	}
	// Since a volume can only be attached to one server
	// we can use the volume id as volume attachment id.
	d.SetId(strconv.Itoa(volume.ID))
	return resourceVolumeAttachmentRead(d, m)
}

func resourceVolumeAttachmentRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

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
		return err
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
		return err
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

func resourceVolumeAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	volumeID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	volume, _, err := client.Volume.GetByID(ctx, volumeID)
	if volume == nil {
		log.Printf("[WARN] Volume ID (%v) not found, removing volume attachment from state", d.Get("volume_id"))
		d.SetId("")
		return nil
	}
	if volume.Server != nil {
		action, _, err := client.Volume.Detach(ctx, volume)
		if err != nil {
			if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
				log.Printf("[INFO] Server (%v) locked, retrying in one second", volume.Server.ID)
				time.Sleep(time.Second)
				return resourceVolumeAttachmentDelete(d, m)
			}
			return err
		}
		if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
			return err
		}
	}
	return nil
}
