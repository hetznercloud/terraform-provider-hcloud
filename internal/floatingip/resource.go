package floatingip

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud FloatingIP resource.
const ResourceType = "hcloud_floating_ip"

// Resource creates a new Terraform schema for the hcloud_floating_ip resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFloatingIPCreate,
		ReadContext:   resourceFloatingIPRead,
		UpdateContext: resourceFloatingIPUpdate,
		DeleteContext: resourceFloatingIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"home_location": {
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
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					if ok, err := hcloud.ValidateResourceLabels(i.(map[string]interface{})); !ok {
						return diag.Errorf(err.Error())
					}
					return nil
				},
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceFloatingIPCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	opts := hcloud.FloatingIPCreateOpts{
		Type:        hcloud.FloatingIPType(d.Get("type").(string)),
		Description: hcloud.Ptr(d.Get("description").(string)),
	}
	if name, ok := d.GetOk("name"); ok {
		opts.Name = hcloud.Ptr(name.(string))
	}
	if serverID, ok := d.GetOk("server_id"); ok {
		opts.Server = &hcloud.Server{ID: serverID.(int)}
	}
	if homeLocation, ok := d.GetOk("home_location"); ok {
		opts.HomeLocation = &hcloud.Location{Name: homeLocation.(string)}
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	res, _, err := client.FloatingIP.Create(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	d.SetId(strconv.Itoa(res.FloatingIP.ID))
	if res.Action != nil {
		_, errCh := client.Action.WatchProgress(ctx, res.Action)
		if err := <-errCh; err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	deleteProtection := d.Get("delete_protection").(bool)
	if deleteProtection {
		if err := setProtection(ctx, client, res.FloatingIP, deleteProtection); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	return resourceFloatingIPRead(ctx, d, m)
}

func resourceFloatingIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Floating IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	floatingIP, _, err := client.FloatingIP.GetByID(ctx, id)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if floatingIP == nil {
		log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setFloatingIPSchema(d, floatingIP)
	return nil
}

func resourceFloatingIPUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Floating IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	floatingIP := &hcloud.FloatingIP{ID: id}

	d.Partial(true)

	if d.HasChange("description") {
		description := d.Get("description").(string)
		_, _, err := client.FloatingIP.Update(ctx, floatingIP, hcloud.FloatingIPUpdateOpts{
			Description: description,
		})
		if err != nil {
			if resourceFloatingIPIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, _, err := client.FloatingIP.Update(ctx, floatingIP, hcloud.FloatingIPUpdateOpts{
			Name: name,
		})
		if err != nil {
			if resourceFloatingIPIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("server_id") {
		serverID := d.Get("server_id").(int)
		if serverID == 0 {
			action, _, err := client.FloatingIP.Unassign(ctx, floatingIP)
			if err != nil {
				if resourceFloatingIPIsNotFound(err, d) {
					return nil
				}
				return hcclient.ErrorToDiag(err)
			}
			if err := hcclient.WaitForAction(ctx, &client.Action, action); err != nil {
				return hcclient.ErrorToDiag(err)
			}
		} else {
			a, _, err := client.FloatingIP.Assign(ctx, floatingIP, &hcloud.Server{ID: serverID})
			if err != nil {
				if resourceFloatingIPIsNotFound(err, d) {
					return nil
				}
				return hcclient.ErrorToDiag(err)
			}
			if err := hcclient.WaitForAction(ctx, &client.Action, a); err != nil {
				return hcclient.ErrorToDiag(err)
			}
		}
	}
	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.FloatingIP.Update(ctx, floatingIP, hcloud.FloatingIPUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceFloatingIPIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("delete_protection") {
		deletionProtection := d.Get("delete_protection").(bool)
		if err := setProtection(ctx, client, floatingIP, deletionProtection); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	d.Partial(false)

	return resourceFloatingIPRead(ctx, d, m)
}

func resourceFloatingIPDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	floatingIPID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Floating IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.FloatingIP.Delete(ctx, &hcloud.FloatingIP{ID: floatingIPID}); err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// server has already been deleted
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func resourceFloatingIPIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setFloatingIPSchema(d *schema.ResourceData, f *hcloud.FloatingIP) {
	for key, val := range getFloatingIPAttributes(f) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getFloatingIPAttributes(f *hcloud.FloatingIP) map[string]interface{} {
	res := map[string]interface{}{
		"id":                f.ID,
		"ip_address":        f.IP.String(),
		"name":              f.Name,
		"type":              f.Type,
		"home_location":     f.HomeLocation.Name,
		"description":       f.Description,
		"labels":            f.Labels,
		"delete_protection": f.Protection.Delete,
	}

	if f.Type == hcloud.FloatingIPTypeIPv6 {
		res["ip_network"] = f.Network.String()
	}
	if f.Server != nil {
		res["server_id"] = f.Server.ID
	}

	return res
}

func setProtection(ctx context.Context, c *hcloud.Client, f *hcloud.FloatingIP, delete bool) error {
	action, _, err := c.FloatingIP.ChangeProtection(ctx, f,
		hcloud.FloatingIPChangeProtectionOpts{
			Delete: &delete,
		},
	)
	if err != nil {
		return err
	}

	return hcclient.WaitForAction(ctx, &c.Action, action)
}
