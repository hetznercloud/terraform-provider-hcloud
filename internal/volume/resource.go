package volume

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
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
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics { // nolint:revive
					if ok, err := hcloud.ValidateResourceLabels(i.(map[string]interface{})); !ok {
						return diag.FromErr(err)
					}
					return nil
				},
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
			"delete_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
		opts.Server = &hcloud.Server{ID: util.CastInt64(serverID)}
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
		opts.Automount = hcloud.Ptr(automount.(bool))
	}
	if format, ok := d.GetOk("format"); ok {
		opts.Format = hcloud.Ptr(format.(string))
	}

	result, _, err := c.Volume.Create(ctx, opts)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			return resourceVolumeCreate(ctx, d, m)
		}
		return hcloudutil.ErrorToDiag(err)
	}
	d.SetId(util.FormatID(result.Volume.ID))

	if err = c.Action.WaitFor(ctx, result.Action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	for _, nextAction := range result.NextActions {
		if err = c.Action.WaitFor(ctx, nextAction); err != nil {
			var aerr hcloud.ActionError

			if nextAction.Command != "attach_volume" {
				return hcloudutil.ErrorToDiag(err)
			}
			if !errors.As(err, &aerr) {
				return hcloudutil.ErrorToDiag(err)
			}
			if !strings.Contains(aerr.Message, string(hcloud.ErrorCodeLocked)) {
				return hcloudutil.ErrorToDiag(err)
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
							opts.Automount = hcloud.Ptr(automount.(bool))
						}
						action, _, err := c.Volume.AttachWithOpts(ctx, result.Volume, o)
						if err != nil {
							return err
						}
						return c.Action.WaitFor(ctx, action)
					})
					if err != nil {
						return hcloudutil.ErrorToDiag(err)
					}
				}
			}
		}
	}

	deleteProtection := d.Get("delete_protection").(bool)
	if deleteProtection {
		if err := setProtection(ctx, c, result.Volume, deleteProtection); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	volume, _, err := client.Volume.GetByID(ctx, id)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
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

	id, err := util.ParseID(d.Id())
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
		return hcloudutil.ErrorToDiag(err)
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
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("server_id") {
		serverID := util.CastInt64(d.Get("server_id"))
		if serverID == 0 {
			err := control.Retry(control.DefaultRetries, func() error {
				action, _, err := c.Volume.Detach(ctx, volume)
				if err != nil {
					if resourceVolumeIsNotFound(err, d) {
						return nil
					}
					return err
				}

				return c.Action.WaitFor(ctx, action)
			})
			if err != nil {
				return hcloudutil.ErrorToDiag(err)
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

					return c.Action.WaitFor(ctx, action)
				})
				if err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
			}
			err := control.Retry(control.DefaultRetries, func() error {
				opts := hcloud.VolumeAttachOpts{Server: &hcloud.Server{ID: serverID}}
				if automount, ok := d.GetOk("automount"); ok {
					opts.Automount = hcloud.Ptr(automount.(bool))
				}

				action, _, err := c.Volume.AttachWithOpts(ctx, volume, opts)
				if err != nil {
					if resourceVolumeIsNotFound(err, d) {
						return nil
					}
					return err
				}

				return c.Action.WaitFor(ctx, action)
			})
			if err != nil {
				return hcloudutil.ErrorToDiag(err)
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
			return hcloudutil.ErrorToDiag(err)
		}

		if err = c.Action.WaitFor(ctx, action); err != nil {
			return hcloudutil.ErrorToDiag(err)
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
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("delete_protection") {
		deletionProtection := d.Get("delete_protection").(bool)
		if err := setProtection(ctx, c, volume, deletionProtection); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	d.Partial(false)
	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	volumeID, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid volume id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	volume, _, err := c.Volume.GetByID(ctx, volumeID)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if volume == nil {
		log.Printf("[WARN] volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if volume.Server != nil {
		err := control.Retry(control.DefaultRetries, func() error {
			action, _, err := c.Volume.Detach(ctx, volume)
			if err != nil {
				if resourceVolumeIsNotFound(err, d) {
					return nil
				}
				return err
			}

			return c.Action.WaitFor(ctx, action)
		})
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
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
		return hcloudutil.ErrorToDiag(err)
	}

	return nil
}

func resourceVolumeIsNotFound(err error, d *schema.ResourceData) bool {
	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		log.Printf("[WARN] volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setVolumeSchema(d *schema.ResourceData, v *hcloud.Volume) {
	util.SetSchemaFromAttributes(d, getVolumeAttributes(v))
}

func getVolumeAttributes(v *hcloud.Volume) map[string]interface{} {
	res := map[string]interface{}{
		"id":                v.ID,
		"name":              v.Name,
		"size":              v.Size,
		"location":          v.Location.Name,
		"labels":            v.Labels,
		"linux_device":      v.LinuxDevice,
		"delete_protection": v.Protection.Delete,
	}

	if v.Server != nil {
		res["server_id"] = v.Server.ID
	}

	return res
}

func setProtection(ctx context.Context, c *hcloud.Client, v *hcloud.Volume, deleteProtection bool) error {
	action, _, err := c.Volume.ChangeProtection(ctx, v,
		hcloud.VolumeChangeProtectionOpts{
			Delete: &deleteProtection,
		},
	)
	if err != nil {
		return err
	}

	return c.Action.WaitFor(ctx, action)
}
