package rdns

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud SSH Key resource.
const ResourceType = "hcloud_rdns"

// Resource creates a Terraform schema for the hcloud_rdns resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReverseDNSCreate,
		ReadContext:   resourceReverseDNSRead,
		UpdateContext: resourceReverseDNSUpdate,
		DeleteContext: resourceReverseDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
			},
		},
	}
}

func resourceReverseDNSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	server, floatingIP, ip, err := lookupRDNSID(ctx, d.Id(), client)
	if err == errInvalidRDNSID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil && floatingIP == nil {
		log.Printf("[WARN] RDNS entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if floatingIP != nil {
		dnsPtr := floatingIP.DNSPtrForIP(ip)
		if dnsPtr == "" {
			log.Printf("[WARN] RDNS entry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		d.Set("dns_ptr", dnsPtr)
		d.Set("floating_ip_id", floatingIP.ID)
		d.Set("ip_address", ip.String())
		d.SetId(generateRDNSID(nil, floatingIP, ip.String()))
		return nil
	}

	if ip.To4() != nil {
		d.SetId(generateRDNSID(server, nil, ip.String()))
		d.Set("dns_ptr", server.PublicNet.IPv4.DNSPtr)
		d.Set("server_id", server.ID)
		d.Set("ip_address", server.PublicNet.IPv4.IP.String())
		return nil
	}

	for rdns := range server.PublicNet.IPv6.DNSPtr {
		if rdns == ip.String() {
			d.SetId(generateRDNSID(server, nil, ip.String()))
			d.Set("dns_ptr", server.PublicNet.IPv6.DNSPtrForIP(ip))
			d.Set("ip_address", ip.String())
			d.Set("server_id", server.ID)
			return nil
		}
	}

	log.Printf("[WARN] RDNS entry (%s) not found, removing from state", d.Id())
	d.SetId("")
	return nil
}

func resourceReverseDNSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)
	ip := d.Get("ip_address").(string)
	ptr := d.Get("dns_ptr").(string)

	serverID, serverOK := d.GetOk("server_id")
	floatingIPID, floatingIPOK := d.GetOk("floating_ip_id")
	if serverOK && floatingIPOK {
		log.Printf("[ERR] server_id and floating_ip_id are mutually exclusive, resource %s", d.Id())
		d.SetId("")
		return nil
	}
	if serverOK {
		server, _, err := c.Server.GetByID(ctx, serverID.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if server == nil {
			log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		d.SetId(generateRDNSID(server, nil, ip))
		action, _, err := c.Server.ChangeDNSPtr(ctx, server, ip, &ptr)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
			return hcclient.ErrorToDiag(err)
		}

		return resourceReverseDNSRead(ctx, d, m)
	}
	if !floatingIPOK {
		log.Printf("[WARN] Invalid floating_ip_id (%s), removing from state: %v", d.Id(), floatingIPOK)
		d.SetId("")
		return nil
	}

	floatingIP, _, err := c.FloatingIP.GetByID(ctx, floatingIPID.(int))
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if floatingIP == nil {
		log.Printf("[WARN] Floating IP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(generateRDNSID(nil, floatingIP, ip))
	action, _, err := c.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, &ptr)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return resourceReverseDNSRead(ctx, d, m)
}

func resourceReverseDNSUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	server, floatingIP, _, err := lookupRDNSID(ctx, d.Id(), c)
	if err == errInvalidRDNSID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil && floatingIP == nil {
		log.Printf("[WARN] RDNS entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	ip := d.Get("ip_address").(string)
	ptr := d.Get("dns_ptr").(string)

	if d.HasChange("dns_ptr") {
		if floatingIP != nil {
			action, _, err := c.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip, &ptr)
			if err != nil {
				return hcclient.ErrorToDiag(err)
			}
			if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
				return hcclient.ErrorToDiag(err)
			}
		} else if server != nil {
			action, _, err := c.Server.ChangeDNSPtr(ctx, server, ip, &ptr)
			if err != nil {
				return hcclient.ErrorToDiag(err)
			}
			if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
				return hcclient.ErrorToDiag(err)
			}
		}
	}
	return resourceReverseDNSRead(ctx, d, m)
}

func resourceReverseDNSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	server, floatingIP, ip, err := lookupRDNSID(ctx, d.Id(), c)
	if err == errInvalidRDNSID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil && floatingIP == nil {
		log.Printf("[WARN] RDNS entry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if floatingIP != nil {
		action, _, err := c.FloatingIP.ChangeDNSPtr(ctx, floatingIP, ip.String(), nil)
		if err != nil {
			if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
				// floating ip has already been deleted
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
		if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
			return hcclient.ErrorToDiag(err)
		}
		return nil
	}

	action, _, err := c.Server.ChangeDNSPtr(ctx, server, ip.String(), nil)
	if err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// server has already been deleted
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
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

var errInvalidRDNSID = errors.New("invalid rdns id")

// lookupRDNSID parses the terraform rdns record id and returns the associated server or floating ip
//
// id format: <prefix>-<resource id>-<ip address>
// Examples:
// s-123-192.168.100.1 (reverse dns entry on server with id 123, ip 192.168.100.1)
// f-123-2001:db8::1 (reverse dns entry on floating ip with id 123, ip 2001:db8::1)
func lookupRDNSID(ctx context.Context, terraformID string, client *hcloud.Client) (
	server *hcloud.Server, floatingIP *hcloud.FloatingIP, ip net.IP, err error) {
	if terraformID == "" {
		err = errInvalidRDNSID
		return
	}

	parts := strings.SplitN(terraformID, "-", 3)
	if len(parts) != 3 {
		err = errInvalidRDNSID
		return
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		err = errInvalidRDNSID
		return
	}

	ip = net.ParseIP(parts[2])
	if ip == nil {
		err = errInvalidRDNSID
		return
	}

	switch parts[0] {
	case "s":
		server, _, err = client.Server.GetByID(ctx, id)
	case "f":
		floatingIP, _, err = client.FloatingIP.GetByID(ctx, id)
	default:
		err = errInvalidRDNSID
	}
	return
}
