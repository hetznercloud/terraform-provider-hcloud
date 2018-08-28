package hcloud

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceFloatingIPAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceFloatingIPAssociationCreate,
		Read:   resourceFloatingIPAssociationRead,
		Update: resourceFloatingIPAssociationUpdate,
		Delete: resourceFloatingIPAssociationDelete,

		Schema: map[string]*schema.Schema{
			"floating_ip_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceFloatingIPAssociationCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	floatingIPID, ok := d.GetOk("floating_ip_id")
	if !ok {
		return fmt.Errorf("could not find floating ip id")
	}
	floatingIP := &hcloud.FloatingIP{ID: floatingIPID.(int)}

	serverID, ok := d.GetOk("server_id")
	if !ok {
		return fmt.Errorf("could not find server id")
	}
	server := &hcloud.Server{ID: serverID.(int)}

	_, _, err := client.FloatingIP.Assign(ctx, floatingIP, server)
	if err != nil {
		return err
	}

	// Since a floating ip can only be assigned to one server
	// we can use the floating ip id as floating ip association id.
	d.SetId(strconv.Itoa(floatingIP.ID))
	return resourceFloatingIPAssociationRead(d, m)
}

func resourceFloatingIPAssociationRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	_, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Floating IP Association ID (%s) not found, removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	// 'floating_ip_id' and 'server_id' is 'Required' and 'TypeInt'
	// therefore the cast should always work
	floatingIP, _, err := client.FloatingIP.GetByID(ctx, d.Get("floating_ip_id").(int))
	if err != nil {
		return err
	}
	if floatingIP == nil {
		log.Printf("[WARN] Floating IP ID (%v) not found, removing Floating IP Association from state", d.Get("floating_ip_id"))
		d.SetId("")
		return nil
	}

	server, _, err := client.Server.GetByID(ctx, d.Get("server_id").(int))
	if err != nil {
		return err
	}
	if server == nil {
		log.Printf("[WARN] Server ID (%v) not found, removing Floating IP Association from state", d.Get("server_id"))
		d.SetId("")
		return nil
	}

	// check if correct server is associated
	if floatingIP.Server != nil {
		server.ID = floatingIP.Server.ID
	} else {
		log.Printf("[WARN] Floating IP (%v) is not associated to a server, removing Floating IP Association from state", d.Get("floating_ip_id"))
		d.SetId("")
		return nil
	}

	d.Set("server_id", server.ID)
	d.Set("floating_ip_id", floatingIP.ID)
	return nil
}

func resourceFloatingIPAssociationUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	_, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Floating IP Association ID (%s) not found, removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	if d.HasChange("server_id") {
		floatingIPID, ok := d.GetOk("floating_ip_id")
		if !ok {
			log.Printf("[WARN] Floating IP ID (%v) not found, removing Floating IP Association from state", d.Get("floating_ip_id"))
			d.SetId("")
			return nil
		}
		floatingIP := &hcloud.FloatingIP{ID: floatingIPID.(int)}

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
	}

	return resourceFloatingIPAssociationRead(d, m)
}

func resourceFloatingIPAssociationDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	_, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Floating IP Association ID (%s) not found, removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	floatingIPID, ok := d.GetOk("floating_ip_id")
	if !ok {
		log.Printf("[WARN] Floating IP ID (%v) not found, removing Floating IP Association from state", d.Get("floating_ip_id"))
		d.SetId("")
		return nil
	}
	floatingIP := &hcloud.FloatingIP{ID: floatingIPID.(int)}

	_, _, err = client.FloatingIP.Unassign(ctx, floatingIP)
	if err != nil {
		return err
	}

	return nil
}
