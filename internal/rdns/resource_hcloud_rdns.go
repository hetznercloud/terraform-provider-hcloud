package rdns

import (
	"context"
	"fmt"
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
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"server_id", "floating_ip_id", "load_balancer_id"},
			},
			"floating_ip_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"server_id", "floating_ip_id", "load_balancer_id"},
			},
			"load_balancer_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"server_id", "floating_ip_id", "load_balancer_id"},
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
	c := m.(*hcloud.Client)

	rdns, ip, err := lookupRDNSID(ctx, d.Id(), c)
	if err != nil {
		d.SetId("")
		return hcclient.ErrorToDiag(err)
	}
	dns, err := rdns.GetDNSPtrForIP(ip)
	if err != nil {
		d.SetId("")
		return hcclient.ErrorToDiag(err)
	}

	d.SetId(generateRDNSID(rdns, ip))
	d.Set("dns_ptr", dns)
	d.Set("ip_address", ip.String())

	switch v := rdns.(type) {
	case *hcloud.Server:
		d.Set("server_id", v.ID)
	case *hcloud.FloatingIP:
		d.Set("floating_ip_id", v.ID)
	case *hcloud.LoadBalancer:
		d.Set("load_balancer_id", v.ID)
	default:
		d.SetId("")
		return diag.Errorf("RDNS does not support %+v", v)
	}

	return nil
}

func resourceReverseDNSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)
	serverID, serverOK := d.GetOk("server_id")
	floatingIPID, floatingIPOK := d.GetOk("floating_ip_id")
	loadBalancerID, loadBalancerOK := d.GetOk("load_balancer_id")

	ip := net.ParseIP(d.Get("ip_address").(string))
	if ip == nil {
		return hcclient.ErrorToDiag(fmt.Errorf("could not parse ip %s", d.Get("ip_address").(string)))
	}
	ptr := d.Get("dns_ptr").(string)

	var rdns hcloud.RDNSSupporter
	var err error

	switch {
	case serverOK:
		rdns, _, err = c.Server.GetByID(ctx, serverID.(int))
	case floatingIPOK:
		rdns, _, err = c.FloatingIP.GetByID(ctx, floatingIPID.(int))
	case loadBalancerOK:
		rdns, _, err = c.LoadBalancer.GetByID(ctx, loadBalancerID.(int))
	}

	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	d.SetId(generateRDNSID(rdns, ip))
	action, _, err := c.RDNS.ChangeDNSPtr(ctx, rdns, ip, &ptr)
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

	rdns, _, err := lookupRDNSID(ctx, d.Id(), c)
	if err != nil {
		d.SetId("")
		return hcclient.ErrorToDiag(err)
	}

	ip := net.ParseIP(d.Get("ip_address").(string))
	if ip == nil {
		return hcclient.ErrorToDiag(fmt.Errorf("could not parse ip %s", d.Get("ip_address").(string)))
	}
	ptr := d.Get("dns_ptr").(string)

	if d.HasChange("dns_ptr") {
		action, _, err := c.RDNS.ChangeDNSPtr(ctx, rdns, ip, &ptr)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}

		if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	return resourceReverseDNSRead(ctx, d, m)
}

func resourceReverseDNSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	rdns, ip, err := lookupRDNSID(ctx, d.Id(), c)
	if err != nil {
		d.SetId("")
		return hcclient.ErrorToDiag(err)
	}

	action, _, err := c.RDNS.ChangeDNSPtr(ctx, rdns, ip, nil)

	if err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// resource has already been deleted
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func generateRDNSID(rdns hcloud.RDNSSupporter, ip net.IP) string {
	switch v := rdns.(type) {
	case *hcloud.Server:
		return fmt.Sprintf("s-%d-%s", v.ID, ip)
	case *hcloud.FloatingIP:
		return fmt.Sprintf("f-%d-%s", v.ID, ip)
	case *hcloud.LoadBalancer:
		return fmt.Sprintf("l-%d-%s", v.ID, ip)
	default:
		return ""
	}
}

type InvalidRDNSIDError struct {
	ID string
}

func (e InvalidRDNSIDError) Error() string {
	return fmt.Sprintf("invalid rdns id %s", e.ID)
}

// lookupRDNSID parses the terraform rdns record id and returns the associated server or floating ip
//
// id format: <prefix>-<resource id>-<ip address>
// Examples:
// s-123-192.168.100.1 (reverse dns entry on server with id 123, ip 192.168.100.1)
// f-123-2001:db8::1 (reverse dns entry on floating ip with id 123, ip 2001:db8::1)
// l-123-2001:db8::1 (reverse dns entry on load balancer with id 123, ip 2001:db8::1)
func lookupRDNSID(ctx context.Context, terraformID string, client *hcloud.Client) (hcloud.RDNSSupporter, net.IP, error) {
	if terraformID == "" {
		return nil, nil, InvalidRDNSIDError{terraformID}
	}

	parts := strings.SplitN(terraformID, "-", 3)
	if len(parts) != 3 {
		return nil, nil, InvalidRDNSIDError{terraformID}
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil, InvalidRDNSIDError{terraformID}
	}

	ip := net.ParseIP(parts[2])
	if ip == nil {
		return nil, nil, InvalidRDNSIDError{terraformID}
	}

	var rdns hcloud.RDNSSupporter

	switch parts[0] {
	case "s":
		rdns, _, err = client.Server.GetByID(ctx, id)
	case "f":
		rdns, _, err = client.FloatingIP.GetByID(ctx, id)
	case "l":
		rdns, _, err = client.LoadBalancer.GetByID(ctx, id)
	default:
		err = InvalidRDNSIDError{terraformID}
	}

	return rdns, ip, err
}
