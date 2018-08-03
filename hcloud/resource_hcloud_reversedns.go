package hcloud

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceReverseDNS() *schema.Resource {
	return &schema.Resource{
		Read:   resourceReverseDNSRead,
		Create: resourceReverseDNSCreate,
		Delete: resourceReverseDNSDelete,

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

func resourceReverseDNSRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	ipAddress := string(d.Get("ip_address").(string))
	rdnsType := ""
	id, ok := d.GetOk("server_id")
	if ok == false {
		id, ok = d.GetOk("floating_ip_id")
		switch {
		case ok == false:
			log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), ok)
			d.SetId("")
			return nil
		case ok == true:
			rdnsType = "floating_ip"
		}
	} else {
		rdnsType = "server"
	}
	ip := net.ParseIP(d.Get("ip_address").(string))
	if ip == nil {
		log.Printf("[WARN] The ip you provide (%s) is not valid, removing from state", ipAddress)
		d.SetId("")
		return nil
	}
	switch {
	case rdnsType == "floating_ip":
		floatingIP, _, err := client.FloatingIP.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if floatingIP == nil {
			log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		switch floatingIP.Type {
		case "ipv4":
			if strings.Count(ipAddress, ".") != 3 {
				log.Printf("[WARN] Floating IP (%s) is an ipv4 floating ip but you dont write a valid one, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		case "ipv6":
			if strings.Count(ipAddress, ":") > 1 {
				log.Printf("[WARN] Floating IP (%s) is an ipv6 but you write an invalid ip, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		}
		d.Set("dns_ptr", floatingIP.DNSPtrForIP(ip))
		d.SetId(fmt.Sprintf("f-%d-"+ipAddress, id))
	case rdnsType == "server":
		server, _, err := client.Server.GetByID(ctx, id.(int))

		if err != nil {
			return err
		}
		if server == nil {
			log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if strings.Count(ipAddress, ".") != 4 {
			d.SetId(fmt.Sprintf("s-%d-"+ipAddress, id))
			d.Set("dns_ptr", server.PublicNet.IPv4.DNSPtr)
		} else if strings.Count(ipAddress, ":") > 1 {
			for rdns := range server.PublicNet.IPv6.DNSPtr {
				if rdns == ipAddress {
					d.SetId(fmt.Sprintf("s-%d-"+ipAddress, id))
					d.Set("dns_ptr", server.PublicNet.IPv6.DNSPtrForIP(ip))
				}
			}
		}

	}
	return nil
}

func resourceReverseDNSCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	ip := d.Get("ip_address").(string)
	ptr := d.Get("dns_ptr").(string)

	rdnsType := ""

	id, ok := d.GetOk("server_id")
	if ok == false {
		id, ok = d.GetOk("floating_ip_id")
		switch {
		case ok == false:
			log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), ok)
			d.SetId("")
			return nil
		}
		d.SetId(fmt.Sprintf("f-%d-"+ip, id))
		rdnsType = "floating_ip"
	} else {
		d.SetId(fmt.Sprintf("s-%d-"+ip, id))
		rdnsType = "server"
	}

	switch {
	case rdnsType == "floating_ip":
		floatingIP, _, err := client.FloatingIP.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if floatingIP == nil {
			log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		action, _, err := client.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, &ptr)

		if err := waitForFloatingIPAction(ctx, client, action, floatingIP); err != nil {
			return err
		}
	case rdnsType == "server":
		server, _, err := client.Server.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if server == nil {
			log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		action, _, err := client.Server.ChangeDNSPtr(ctx, server, ip, &ptr)

		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}

	}
	return resourceReverseDNSRead(d, m)
}
func resourceReverseDNSDelete(d *schema.ResourceData, m interface{}) error {
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
