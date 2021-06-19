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

// AssignmentResourceType is the type name of the Hetzner Cloud FloatingIP resource.
const AssignmentResourceType = "hcloud_floating_ip_assignment"

// AssignmentResource creates a new Terraform schema for the
// hcloud_floating_ip_assignment resource.
func AssignmentResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFloatingIPAssignmentCreate,
		ReadContext:   resourceFloatingIPAssignmentRead,
		DeleteContext: resourceFloatingIPAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"floating_ip_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFloatingIPAssignmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	floatingIPID := d.Get("floating_ip_id")
	floatingIP := &hcloud.FloatingIP{ID: floatingIPID.(int)}

	serverID := d.Get("server_id")

	server := &hcloud.Server{ID: serverID.(int)}

	_, _, err := client.FloatingIP.Assign(ctx, floatingIP, server)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	// Since a floating ip can only be assigned to one server
	// we can use the floating ip id as floating ip association id.
	d.SetId(strconv.Itoa(floatingIP.ID))
	return resourceFloatingIPAssignmentRead(ctx, d, m)
}

func resourceFloatingIPAssignmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	floatingIPID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Floating IP ID (%s) not found, removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	// 'floating_ip_id' and 'server_id' is 'Required' and 'TypeInt'
	// therefore the cast should always work
	floatingIP, _, err := client.FloatingIP.GetByID(ctx, floatingIPID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if floatingIP == nil {
		log.Printf("[WARN] Floating IP ID (%v) not found, removing Floating IP Association from state", d.Get("floating_ip_id"))
		d.SetId("")
		return nil
	}

	// check if floating api is assigned to any server
	if floatingIP.Server == nil {
		log.Printf("[WARN] Floating IP (%v) is not associated to a server, removing Floating IP Association from state", d.Get("floating_ip_id"))
		d.SetId("")
		return nil
	}

	// when importing the resource the server_id is not given
	// because only the terraform ID (Floating IP ID in this case)
	// is available, so we need to get the ID from the volume
	// instead of from the configuration
	serverID := d.Get("server_id").(int)
	if serverID == 0 {
		serverID = floatingIP.Server.ID
	}
	server, _, err := client.Server.GetByID(ctx, serverID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil {
		log.Printf("[WARN] Server ID (%v) not found, removing Floating IP Association from state", d.Get("server_id"))
		d.SetId("")
		return nil
	}

	d.Set("server_id", floatingIP.Server.ID)
	d.Set("floating_ip_id", floatingIP.ID)
	return nil
}

func resourceFloatingIPAssignmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	floatingIPID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	floatingIP, _, err := client.FloatingIP.GetByID(ctx, floatingIPID)
	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if floatingIP == nil {
		log.Printf("[WARN] Floating IP ID (%v) not found, removing Floating IP Association from state", d.Get("floating_ip_id"))
		d.SetId("")
		return nil
	}
	if floatingIP.Server != nil {
		_, _, err = client.FloatingIP.Unassign(ctx, floatingIP)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	return nil
}
