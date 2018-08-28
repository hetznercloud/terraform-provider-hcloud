package hcloud

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceFloatingIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceFloatingIPCreate,
		Read:   resourceFloatingIPRead,
		Update: resourceFloatingIPUpdate,
		Delete: resourceFloatingIPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
		},
	}
}

func resourceFloatingIPCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	opts := hcloud.FloatingIPCreateOpts{
		Type:        hcloud.FloatingIPType(d.Get("type").(string)),
		Description: hcloud.String(d.Get("description").(string)),
	}

	if serverID, ok := d.GetOk("server_id"); ok {
		opts.Server = &hcloud.Server{ID: serverID.(int)}
	}
	if homeLocation, ok := d.GetOk("home_location"); ok {
		opts.HomeLocation = &hcloud.Location{Name: homeLocation.(string)}
	}

	res, _, err := client.FloatingIP.Create(ctx, opts)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(res.FloatingIP.ID))
	if res.Action != nil {
		_, errCh := client.Action.WatchProgress(ctx, res.Action)
		if err := <-errCh; err != nil {
			return err
		}
	}

	return resourceFloatingIPRead(d, m)
}

func resourceFloatingIPRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Floating IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	floatingIP, _, err := client.FloatingIP.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if floatingIP == nil {
		log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("description", floatingIP.Description)
	d.Set("home_location", floatingIP.HomeLocation.Name)
	d.Set("type", floatingIP.Type)
	if floatingIP.Server != nil {
		d.Set("server_id", floatingIP.Server.ID)
	}
	d.Set("ip_address", floatingIP.IP.String())
	if floatingIP.Type == hcloud.FloatingIPTypeIPv6 {
		d.Set("ip_network", floatingIP.Network.String())
	}
	return nil
}

func resourceFloatingIPUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

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
			return err
		}
		d.SetPartial("description")
	}

	if d.HasChange("server_id") {
		serverID := d.Get("server_id").(int)
		if serverID == 0 {
			action, _, err := client.FloatingIP.Unassign(ctx, floatingIP)
			if err != nil {
				if resourceFloatingIPIsNotFound(err, d) {
					return nil
				}
				return err
			}
			if err := waitForFloatingIPAction(ctx, client, action, floatingIP); err != nil {
				return err
			}
		} else {
			action, _, err := client.FloatingIP.Assign(ctx, floatingIP, &hcloud.Server{ID: serverID})
			if err != nil {
				if resourceFloatingIPIsNotFound(err, d) {
					return nil
				}
				return err
			}
			if err := waitForFloatingIPAction(ctx, client, action, floatingIP); err != nil {
				return err
			}
		}
		d.SetPartial("server_id")
	}

	d.Partial(false)

	return resourceFloatingIPRead(d, m)
}

func resourceFloatingIPDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

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
		return err
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

func waitForFloatingIPAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action, floatingIP *hcloud.FloatingIP) error {
	log.Printf("[INFO] Floating IP (%d) waiting for %q action to complete...", floatingIP.ID, action.Command)
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] Floating IP (%d) %q action succeeded", floatingIP.ID, action.Command)
	return nil
}
