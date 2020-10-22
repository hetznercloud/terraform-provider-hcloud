package hcloud

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVolumeCreate,
		ReadContext:   resourceVolumeRead,
		UpdateContext: resourceVolumeUpdate,
		DeleteContext: resourceVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"linux_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automount": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"format": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

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
	if automount, ok := d.GetOk("automount"); ok {
		opts.Automount = hcloud.Bool(automount.(bool))
	}
	if format, ok := d.GetOk("format"); ok {
		opts.Format = hcloud.String(format.(string))
	}

	result, _, err := client.Volume.Create(ctx, opts)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := waitForVolumeAction(ctx, client, result.Action, result.Volume); err != nil {
		return diag.FromErr(err)
	}
	for _, nextAction := range result.NextActions {
		if err := waitForVolumeAction(ctx, client, nextAction, result.Volume); err != nil {
			return diag.FromErr(err)
		}

	}
	d.SetId(strconv.Itoa(result.Volume.ID))

	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	volume, _, err := client.Volume.GetByID(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if volume == nil {
		log.Printf("[WARN] volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setVolumeSchema(d, volume)
	return nil
}

func resourceVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	volume, _, err := client.Volume.GetByID(ctx, id)
	if err != nil {
		if resourceVolumeIsNotFound(err, d) {
			return nil
		}
		return diag.FromErr(err)
	}

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
			return diag.FromErr(err)
		}
	}

	if d.HasChange("server_id") {
		serverID := d.Get("server_id").(int)
		if serverID == 0 {
			action, _, err := client.Volume.Detach(ctx, volume)
			if err != nil {
				if resourceVolumeIsNotFound(err, d) {
					return nil
				}
				return diag.FromErr(err)
			}
			if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if volume.Server != nil {
				action, _, err := client.Volume.Detach(ctx, volume)
				if err != nil {
					if resourceVolumeIsNotFound(err, d) {
						return nil
					}
					return diag.FromErr(err)
				}
				if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
					return diag.FromErr(err)
				}
			}
			action, _, err := client.Volume.Attach(ctx, volume, &hcloud.Server{ID: serverID})

			if err != nil {
				if resourceVolumeIsNotFound(err, d) {
					return nil
				}
				return diag.FromErr(err)
			}
			if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("size") {
		size := d.Get("size").(int)
		action, _, err := client.Volume.Resize(ctx, volume, size)
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return diag.FromErr(err)
		}
		if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
			return diag.FromErr(err)
		}

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
			return diag.FromErr(err)
		}
	}
	d.Partial(false)

	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	volumeID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	volume, _, err := client.Volume.GetByID(ctx, volumeID)
	if err != nil {
		return diag.FromErr(err)
	}

	if volume.Server != nil {
		action, _, err := client.Volume.Detach(ctx, volume)
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return diag.FromErr(err)
		}
		if err := waitForVolumeAction(ctx, client, action, volume); err != nil {
			return diag.FromErr(err)
		}
	}
	if _, err := client.Volume.Delete(ctx, volume); err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// volume has already been deleted
			return nil
		}
		return diag.FromErr(err)
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
	d.Set("location", v.Location.Name)
	if v.Server != nil {
		d.Set("server_id", v.Server.ID)
	}
	d.Set("labels", v.Labels)
	d.Set("linux_device", v.LinuxDevice)
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
