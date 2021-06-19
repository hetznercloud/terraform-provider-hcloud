package floatingip

import (
	"context"
	"log"
	"strconv"

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
			},
		},
	}
}

func resourceFloatingIPCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	opts := hcloud.FloatingIPCreateOpts{
		Type:        hcloud.FloatingIPType(d.Get("type").(string)),
		Description: hcloud.String(d.Get("description").(string)),
	}
	if name, ok := d.GetOk("name"); ok {
		opts.Name = hcloud.String(name.(string))
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
	d.SetId(strconv.Itoa(f.ID))
	d.Set("ip_address", f.IP.String())
	d.Set("name", f.Name)
	if f.Type == hcloud.FloatingIPTypeIPv6 {
		d.Set("ip_network", f.Network.String())
	}
	if f.Server != nil {
		d.Set("server_id", f.Server.ID)
	}
	d.Set("type", f.Type)
	d.Set("home_location", f.HomeLocation.Name)
	d.Set("description", f.Description)
	d.Set("labels", f.Labels)
}
