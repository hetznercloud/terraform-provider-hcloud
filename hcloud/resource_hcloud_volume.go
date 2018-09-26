package hcloud

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceVolumeCreate,
		Read:   resourceVolumeRead,
		Update: resourceVolumeUpdate,
		Delete: resourceVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceVolumeCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	opts := hcloud.VolumeCreateOpts{
		Name: d.Get("name").(string),
		Size: d.Get("size").(int),
	}

	if serverID, ok := d.GetOk("server_id"); ok {
		opts.Server = &hcloud.Server{ID: serverID.(int)}
	}
	if location, ok := d.GetOk("location"); ok {
		opts.Location = &hcloud.Location{Name: location.(string)}
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	result, _, err := client.Volume.Create(ctx, opts)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(result.Volume.ID))

	return resourceVolumeRead(d, m)
}

func resourceVolumeRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	volume, _, err := client.Volume.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if volume == nil {
		log.Printf("[WARN] volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setVolumeSchema(d, volume)
	return nil
}

func resourceVolumeUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	volume := &hcloud.Volume{ID: id}

	d.Partial(true)

	if d.HasChange("name") {
		description := d.Get("name").(string)
		_, _, err := client.Volume.Update(ctx, volume, hcloud.VolumeUpdateOpts{
			Name: description,
		})
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return err
		}
		d.SetPartial("name")
	}

	if d.HasChange("server_id") {
		serverID := d.Get("server_id").(int)
		if serverID == 0 {
			action, _, err := client.Volume.Detach(ctx, volume)
			if err != nil {
				if resourceVolumeIsNotFound(err, d) {
					return nil
				}
				return err
			}
			if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
				return err
			}
		} else {
			action, _, err := client.Volume.Attach(ctx, volume, &hcloud.Server{ID: serverID})
			if err != nil {
				if resourceVolumeIsNotFound(err, d) {
					return nil
				}
				return err
			}
			if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
				return err
			}
		}
		d.SetPartial("server_id")
	}

	if d.HasChange("size") {
		size := d.Get("size").(int)
		action, _, err := client.Volume.Resize(ctx, volume, size)
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return err
		}
		if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
			return err
		}

		d.SetPartial("size")
	}

	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.Volume.Update(ctx, volume, hcloud.VolumeUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return err
		}
		d.SetPartial("labels")
	}
	d.Partial(false)

	return resourceVolumeRead(d, m)
}

func resourceVolumeDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	volumeID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.Volume.Delete(ctx, &hcloud.Volume{ID: volumeID}); err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// server has already been deleted
			return nil
		}
		return err
	}

	return nil
}

func resourceVolumeIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setVolumeSchema(d *schema.ResourceData, v *hcloud.Volume) {
	d.SetId(strconv.Itoa(v.ID))
	d.Set("name", v.Name)
	d.Set("size", v.Size)
	if v.Server != nil {
		d.Set("server_id", v.Server)
	}
	d.Set("labels", v.Labels)
}

func waitForVolumeAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action, volume *hcloud.Volume) error {
	log.Printf("[INFO] volume (%d) waiting for %q action to complete...", volume.ID, action.Command)
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] volume (%d) %q action succeeded", volume.ID, action.Command)
	return nil
}
