package hcloud

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"log"
	"net"
	"strings"
)

func resourceReverseDns() *schema.Resource {
	return &schema.Resource{
		Read:   resourceReverseDnsRead,
		Create: resourceReverseDnsCreate,
		Delete: resourceReverseDnsDelete,

		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"floating_ip_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
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

	id, err := d.Get("server_id").(int)
	rdnsType := ""
	if err != false {
		id, err = d.Get("floating_ip_id").(int)
		if err != false {
			log.Printf("[WARN] invalid id (%s), removing from state: %v", d.Id(), err)
			d.SetId("")
			return nil
		} else {

			rdnsType = "floating_ip"
		}
	} else {
		rdnsType = "server"
	}

	ipAddress := string(d.Get("ip_address").(string))
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
		if floatingIP.Type == "ipv4" {
			if strings.Count(ipAddress, ".") != 4 {
				log.Printf("[WARN] Floating IP (%s) is an ipv4 but you write an ipv6, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		} else if floatingIP.Type == "ipv6" {
			if strings.Count(ipAddress, ":") > 1 {
				log.Printf("[WARN] Floating IP (%s) is an ipv6 but you write an invalid ip, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		}
		d.SetId("f-" + string(id) + "-" + ipAddress)
		d.Set("dns_ptr", floatingIP.DNSPtrForIP(d.Get("ip_address").(net.IP)))
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
		if strings.Count(ipAddress, ".") != 4 {
			d.SetId("s-" + string(id) + "-" + ipAddress)
			d.Set("dns_ptr", server.PublicNet.IPv4.DNSPtr)
		} else if strings.Count(ipAddress, ":") > 1 {
			for rdns := range server.PublicNet.IPv6.DNSPtr {
				if rdns == ipAddress {
					d.SetId("s-" + string(id) + "-" + ipAddress)
					d.Set("dns_ptr", server.PublicNet.IPv6.DNSPtrForIP(d.Get("ip_address").(net.IP)))
				}
			}
		}

	}
	return nil
}

func resourceReverseDnsCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	id, err := d.Get("server_id").(int)
	rdnsType := ""
	if err != false {
		id, err = d.Get("floating_ip_id").(int)
		if err != false {
			log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), err)
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
		ip := d.Get("ip_address").(string)
		ptr := d.Get("dns_ptr").(string)
		d.SetId("f-" + string(id) + "-" + ip)
		action, _, err := client.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, &ptr)

		if err := waitForFloatingIPAction(ctx, client, action, floatingIP); err != nil {
			return err
		}
	} else if rdnsType == "server" {
		server, _, err := client.Server.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if server == nil {
			log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		ip := d.Get("ip_address").(string)
		ptr := d.Get("dns_ptr").(string)
		d.SetId("s-" + string(id) + "-" + ip)
		action, _, err := client.Server.ChangeDNSPtr(ctx, server, ip, &ptr)

		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}

	}
	return resourceReverseDnsRead(d, m)
}
func resourceReverseDnsDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	id, err := d.Get("server_id").(int)
	rdnsType := ""
	if err != false {
		id, err = d.Get("floating_ip_id").(int)
		if err != false {
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
			action, _, err := client.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, nil)

			if err := waitForFloatingIPAction(ctx, client, action, floatingIP); err != nil {
				return err
			}
		}
	} else if rdnsType == "server" {
		server, _, err := client.Server.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if server == nil {
			log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		ip := d.Get("ip_address").(string)
		action, _, err := client.Server.ChangeDNSPtr(ctx, server, ip, nil)

		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}

	return nil
}
