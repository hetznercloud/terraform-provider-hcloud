package hcloud

import (
	"context"
	"fmt"
	"log"
	"net"

	"strconv"
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
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceReverseDNSRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ipAddress := string(d.Get("ip_address").(string))
	server, floatingIP, err := lookupRDNSID(d.Id(), client)

	switch {
	case err != nil:
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	ip := net.ParseIP(d.Get("ip_address").(string))
	if ip == nil {
		log.Printf("[ERR] The ip you provide (%s) is not valid", ipAddress)
		return nil
	}
	switch {
	case floatingIP != nil:
		if floatingIP == nil {
			log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		ip = net.ParseIP(ipAddress)
		switch floatingIP.Type {
		case "ipv4":
			if ip.To4() == nil {
				log.Printf("[WARN] Floating IP (%s) is an ipv4 floating ip but you dont write a valid one, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		case "ipv6":
			if ip.To16() == nil {
				log.Printf("[WARN] Floating IP (%s) is an ipv6 but you write an invalid ip, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		}
		dnsPtr := floatingIP.DNSPtrForIP(ip)
		if dnsPtr != "" {
			d.Set("dnsPtr", dnsPtr)
			d.SetId(generateRDNSID(nil, floatingIP, ipAddress))
		} else {
			d.SetId("")
		}
	case server != nil:
		ip = net.ParseIP(ipAddress)
		if ip.To4() != nil {
			d.SetId(generateRDNSID(server, nil, ipAddress))
			d.Set("dnsPtr", server.PublicNet.IPv4.DNSPtr)
		} else if ip.To16() != nil {
			for rdns := range server.PublicNet.IPv6.DNSPtr {
				if rdns == ipAddress {
					d.SetId(generateRDNSID(server, nil, ipAddress))
					d.Set("dnsPtr", server.PublicNet.IPv6.DNSPtrForIP(ip))
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
		case !ok:
			log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), ok)
			d.SetId("")
			return nil
		}
		floatingIP, _, _ := client.FloatingIP.GetByID(ctx, id.(int))
		d.SetId(generateRDNSID(nil, floatingIP, ip))
		rdnsType = "floatingIP"
	} else {
		server, _, _ := client.Server.GetByID(ctx, id.(int))
		d.SetId(generateRDNSID(server, nil, ip))
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
	server, floatingIP, err := lookupRDNSID(d.Id(), client)
	switch {
	case err != nil:
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if floatingIP != nil {
		if d.HasChange("dns_ptr") {
			ip := d.Get("ip_address").(string)
			action, _, _ := client.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, nil)

			if err := waitForFloatingIPAction(ctx, client, action, floatingIP); err != nil {
				return err
			}
		}
	} else if server != nil {
		ip := d.Get("ip_address").(string)
		action, _, _ := client.Server.ChangeDNSPtr(ctx, server, ip, nil)

		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}

	return nil
}
func generateRDNSID(server *hcloud.Server, floatingIP *hcloud.FloatingIP, ip string) string {
	if server != nil {
		return fmt.Sprintf("s-%d-%s", server.ID, ip)
	}
	if floatingIP != nil {
		return fmt.Sprintf("f-%d-%s", floatingIP.ID, ip)
	}
	return ""
}
func lookupRDNSID(id string, client *hcloud.Client) (server *hcloud.Server, floatingIP *hcloud.FloatingIP, err error) {

	ctx := context.Background()
	ids := strings.Split(id, "-")
	_id, _ := strconv.Atoi(string(ids[1]))
	switch ids[0] {
	case "s":
		server, _, err = client.Server.GetByID(ctx, _id)
	case "f":
		floatingIP, _, err = client.FloatingIP.GetByID(ctx, _id)
	default:
		err = fmt.Errorf("Can not lookup the id, type is unknown")
	}
	return server, floatingIP, err
}
