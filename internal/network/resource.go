package network

import (
	"context"
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud Network resource.
const ResourceType = "hcloud_network"

// Resource creates a Terraform schema for the hcloud_network resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkCreate,
		ReadContext:   resourceNetworkRead,
		UpdateContext: resourceNetworkUpdate,
		DeleteContext: resourceNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_range": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					if ok, err := hcloud.ValidateResourceLabels(i.(map[string]interface{})); !ok {
						return diag.Errorf(err.Error())
					}
					return nil
				},
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	_, ipRange, err := net.ParseCIDR(d.Get("ip_range").(string))
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	opts := hcloud.NetworkCreateOpts{
		Name:    d.Get("name").(string),
		IPRange: ipRange,
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	network, _, err := client.Network.Create(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	d.SetId(strconv.Itoa(network.ID))

	deleteProtection := d.Get("delete_protection").(bool)
	if deleteProtection {
		if err := setProtection(ctx, client, network, deleteProtection); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	return resourceNetworkRead(ctx, d, m)
}

func resourceNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	network, _, err := client.Network.Get(ctx, d.Id())
	if err != nil {
		if resourceNetworkIsNotFound(err, d) {
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}
	if network == nil {
		d.SetId("")
		return nil
	}
	setNetworkSchema(d, network)
	return nil
}

func resourceNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	network, _, err := client.Network.Get(ctx, d.Id())
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if network == nil {
		d.SetId("")
		return nil
	}

	d.Partial(true)
	if d.HasChange("name") {
		newName := d.Get("name")
		_, _, err := client.Network.Update(ctx, network, hcloud.NetworkUpdateOpts{
			Name: newName.(string),
		})
		if err != nil {
			if resourceNetworkIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}
	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.Network.Update(ctx, network, hcloud.NetworkUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceNetworkIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("delete_protection") {
		deletionProtection := d.Get("delete_protection").(bool)
		if err := setProtection(ctx, client, network, deletionProtection); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	d.Partial(false)

	return resourceNetworkRead(ctx, d, m)
}

func resourceNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	networkID, err := strconv.Atoi(d.Id())

	if err != nil {
		log.Printf("[WARN] invalid network id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.Network.Delete(ctx, &hcloud.Network{ID: networkID}); err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// network has already been deleted
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func resourceNetworkIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setNetworkSchema(d *schema.ResourceData, n *hcloud.Network) {
	for key, val := range getNetworkAttributes(n) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getNetworkAttributes(n *hcloud.Network) map[string]interface{} {
	return map[string]interface{}{
		"id":                n.ID,
		"ip_range":          n.IPRange.String(),
		"name":              n.Name,
		"labels":            n.Labels,
		"delete_protection": n.Protection.Delete,
	}
}

func setProtection(ctx context.Context, c *hcloud.Client, n *hcloud.Network, delete bool) error {
	action, _, err := c.Network.ChangeProtection(ctx, n,
		hcloud.NetworkChangeProtectionOpts{
			Delete: &delete,
		},
	)
	if err != nil {
		return err
	}

	return hcclient.WaitForAction(ctx, &c.Action, action)
}
