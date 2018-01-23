package hcloud

import (
	"context"
	"fmt"
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
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reverse_dns": {
				Type:     schema.TypeMap,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
		return fmt.Errorf("invalid floating ip id: %v", err)
	}

	floatingIP, _, err := client.FloatingIP.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if floatingIP == nil {
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
	d.Set("reverse_dns", floatingIP.DNSPtr)

	return nil
}

func resourceFloatingIPUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("invalid floating ip id: %v", err)
	}
	floatingIP := &hcloud.FloatingIP{ID: id}

	d.Partial(true)

	if d.HasChange("description") {
		description := d.Get("description").(string)
		_, _, err := client.FloatingIP.Update(ctx, floatingIP, hcloud.FloatingIPUpdateOpts{
			Description: description,
		})
		if err != nil {
			return err
		}
		d.SetPartial("description")
	}

	if d.HasChange("server_id") {
		serverID := d.Get("server_id").(int)
		if serverID == 0 {
			_, _, err := client.FloatingIP.Unassign(ctx, floatingIP)
			if err != nil {
				return err
			}
		} else {
			_, _, err := client.FloatingIP.Assign(ctx, floatingIP, &hcloud.Server{ID: serverID})
			if err != nil {
				return err
			}
		}
		d.SetPartial("server_id")
	}

	if d.HasChange("reverse_dns") {
		oldPtrRaw, newPtrRaw := d.GetChange("reverse_dns")
		oldPtr := getFloatingIPDNSPtr(oldPtrRaw)
		newPtr := getFloatingIPDNSPtr(newPtrRaw)

		for ip := range oldPtr {
			if _, ok := newPtr[ip]; !ok {
				newPtr[ip] = nil
			}
		}

		if err := updateFloatingIPDNSPtr(ctx, client, floatingIP, newPtr); err != nil {
			return err
		}
		d.SetPartial("reverse_dns")
	}

	d.Partial(false)

	return nil
}

func resourceFloatingIPDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	floatingIPID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Invalid floatingIP ID: %s", d.Id())
	}
	if _, err := client.FloatingIP.Delete(ctx, &hcloud.FloatingIP{ID: floatingIPID}); err != nil {
		return err
	}

	return nil
}

func getFloatingIPDNSPtr(raw interface{}) map[string]*string {
	out := map[string]*string{}
	for k, v := range raw.(map[string]interface{}) {
		out[k] = hcloud.String(v.(string))
	}
	return out
}

func updateFloatingIPDNSPtr(ctx context.Context, client *hcloud.Client, floatingIP *hcloud.FloatingIP, update map[string]*string) error {
	for ip, ptr := range update {
		action, _, err := client.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, ptr)
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}
