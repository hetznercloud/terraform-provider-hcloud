package hcloud

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceReverseDns() *schema.Resource {
	return &schema.Resource{
		Read:   resourceReverseDnsRead,
		Update: resourceReverseDnsUpdate,

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"floating_ip_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"dns_ptr": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func resourceReverseDnsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	id, err := strconv.Atoi(d.Get("server_id").(string))
	rdnsType := ""
	if err != nil {
		id, err = strconv.Atoi(d.Get("floating_ip_id").(string))
		if err != nil {
			log.Printf("[WARN] invalid id (%s), removing from state: %v", d.Id(), err)
			d.SetId("")
			return nil
		} else {
			rdnsType = "floating_ip"
		}
	} else {
		rdnsType = "server"
	}
	if rdnsType == "floating_ip" {
		floatingIP, _, err := client.FloatingIP.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if floatingIP == nil {
			log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		d.Set("ip_address", floatingIP.IP)
		d.Set("dns_ptr", floatingIP.DNSPtrForIP(floatingIP.IP))
	} else if rdnsType == "server" {
		server, _, err := client.Server.Get(ctx, string(id))
		if err != nil {
			return err
		}
		if server == nil {
			log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		d.Set("ip_address", server.PublicNet.IPv4.IP)
		d.Set("dns_ptr", server.PublicNet.IPv4.DNSPtr)
	}
	return nil
}

func resourceReverseDnsUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	id, err := strconv.Atoi(d.Get("server_id").(string))
	rdnsType := ""
	if err != nil {
		id, err = strconv.Atoi(d.Get("floating_ip_id").(string))
		if err != nil {
			log.Printf("[WARN] invalid id (%s), removing from state: %v", d.Id(), err)
			d.SetId("")
			return nil
		} else {
			rdnsType = "floating_ip"
		}
	} else {
		rdnsType = "server"
	}
	d.Partial(true)
	if rdnsType == "floating_ip" {
		if d.HasChange("dns_ptr") {
			floatingIP, _, err := client.FloatingIP.GetByID(ctx, id)
			if err != nil {
				return err
			}
			if floatingIP == nil {
				log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			ip := d.Get("ip_address").(string)
			ptr := d.Get("dns_ptr").(string)
			action, _, err := client.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, &ptr)

			if err := waitForFloatingIPAction(ctx, client, action, floatingIP); err != nil {
				return err
			}
			d.SetPartial("dns_ptr")
		}
	} else if rdnsType == "server" {
		server, _, err := client.Server.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if server == nil {
			log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		ip := d.Get("ip_address").(string)
		ptr := d.Get("dns_ptr").(string)
		action, _, err := client.Server.ChangeDNSPtr(ctx, server, ip, &ptr)

		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
		d.SetPartial("dns_ptr")
	}

	d.Partial(false)

	return resourceReverseDnsRead(d, m)
}
