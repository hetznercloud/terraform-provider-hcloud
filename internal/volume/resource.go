package volume

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud Volume resource.
const ResourceType = "hcloud_volume"

// Resource creates a Terraform schema for the hcloud_volume resource.
func Resource() *schema.Resource {
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
	c := m.(*hcloud.Client)

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

	result, _, err := c.Volume.Create(ctx, opts)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			return resourceVolumeCreate(ctx, d, m)
		}
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, result.Action); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	for _, nextAction := range result.NextActions {
		if err := hcclient.WaitForAction(ctx, &c.Action, nextAction); err != nil {
			var aerr hcloud.ActionError

			if nextAction.Command != "attach_volume" {
				return hcclient.ErrorToDiag(err)
			}
			if !errors.As(err, &aerr) {
				return hcclient.ErrorToDiag(err)
			}
			if !strings.Contains(aerr.Message, string(hcloud.ErrorCodeLocked)) {
				return hcclient.ErrorToDiag(err)
			}

			// Sometimes, when multiple volumes are created (for the same server)
			// we get a failed attach_volume action with the "locked" error identifier
			// We should then simply wait a few seconds and control.Retry the attachment.
			// Use a sleeping timeout that is high enough,
			// most volume actions will be done in less than 5 seconds
			// so with 5 seconds we should be safe enough. after that control.Retry the attach
			// call and it should work.
			for _, resource := range nextAction.Resources {
				if resource.Type == hcloud.ActionResourceTypeServer {
					err := control.Retry(control.DefaultRetries, func() error {
						o := hcloud.VolumeAttachOpts{Server: opts.Server}
						if automount, ok := d.GetOk("automount"); ok {
							opts.Automount = hcloud.Bool(automount.(bool))
						}
						a, _, err := c.Volume.AttachWithOpts(ctx, result.Volume, o)
						if err != nil {
							return err
						}
						return hcclient.WaitForAction(ctx, &c.Action, a)
					})
					if err != nil {
						return hcclient.ErrorToDiag(err)
					}
				}
			}
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
		return hcclient.ErrorToDiag(err)
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
	c := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	volume, _, err := c.Volume.GetByID(ctx, id)
	if err != nil {
		if resourceVolumeIsNotFound(err, d) {
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	d.Partial(true)

	if d.HasChange("name") {
		description := d.Get("name").(string)
		_, _, err := c.Volume.Update(ctx, volume, hcloud.VolumeUpdateOpts{
			Name: description,
		})
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("server_id") {
		serverID := d.Get("server_id").(int)
		if serverID == 0 {
			err := control.Retry(control.DefaultRetries, func() error {
				a, _, err := c.Volume.Detach(ctx, volume)
				if err != nil {
					if resourceVolumeIsNotFound(err, d) {
						return nil
					}
					return err
				}

				if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return hcclient.ErrorToDiag(err)
			}
		} else {
			if volume.Server != nil {
				err := control.Retry(control.DefaultRetries, func() error {
					action, _, err := c.Volume.Detach(ctx, volume)
					if err != nil {
						if resourceVolumeIsNotFound(err, d) {
							return nil
						}
						return err
					}
					if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return hcclient.ErrorToDiag(err)
				}
			}
			err := control.Retry(control.DefaultRetries, func() error {
				opts := hcloud.VolumeAttachOpts{Server: &hcloud.Server{ID: serverID}}
				if automount, ok := d.GetOk("automount"); ok {
					opts.Automount = hcloud.Bool(automount.(bool))
				}

				action, _, err := c.Volume.AttachWithOpts(ctx, volume, opts)
				if err != nil {
					if resourceVolumeIsNotFound(err, d) {
						return nil
					}
					return err
				}
				if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return hcclient.ErrorToDiag(err)
			}
		}
	}

	if d.HasChange("size") {
		size := d.Get("size").(int)
		action, _, err := c.Volume.Resize(ctx, volume, size)
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
		if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := c.Volume.Update(ctx, volume, hcloud.VolumeUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}
	d.Partial(false)

	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	volumeID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	volume, _, err := c.Volume.GetByID(ctx, volumeID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	if volume.Server != nil {
		err := control.Retry(control.DefaultRetries, func() error {
			a, _, err := c.Volume.Detach(ctx, volume)
			if err != nil {
				if resourceVolumeIsNotFound(err, d) {
					return nil
				}
				return err
			}

			if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	err = control.Retry(control.DefaultRetries, func() error {
		if _, err := c.Volume.Delete(ctx, volume); err != nil {
			if resourceVolumeIsNotFound(err, d) {
				return nil
			}
			return err
		}
		return nil
	})
	if err != nil {
		return hcclient.ErrorToDiag(err)
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
