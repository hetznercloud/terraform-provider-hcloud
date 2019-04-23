package hcloud

import (
	"context"
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkCreate,
		Read:   resourceNetworkRead,
		Update: resourceNetworkUpdate,
		Delete: resourceNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_range": {
				Type:     schema.TypeString,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceNetworkCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	_, ipRange, err := net.ParseCIDR(d.Get("ip_range").(string))
	if err != nil {
		return err
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
		return err
	}

	d.SetId(strconv.Itoa(network.ID))

	return resourceNetworkRead(d, m)
}

func resourceNetworkRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	network, _, err := client.Network.Get(ctx, d.Id())
	if err != nil {
		if resourceNetworkIsNotFound(err, d) {
			return nil
		}
		return err
	}
	if network == nil {
		d.SetId("")
		return nil
	}
	setNetworkSchema(d, network)
	return nil

}

func resourceNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	network, _, err := client.Network.Get(ctx, d.Id())
	if err != nil {
		return err
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
			return err
		}
		d.SetPartial("name")
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
			return err
		}
		d.SetPartial("labels")
	}
	d.Partial(false)

	return resourceNetworkRead(d, m)
}

func resourceNetworkDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	networkID, err := strconv.Atoi(d.Id())

	if err != nil {
		log.Printf("[WARN] invalid network id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.Network.Delete(ctx, &hcloud.Network{ID: networkID}); err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// network has already been deleted
			return nil
		}
		return err
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
	d.SetId(strconv.Itoa(n.ID))
	d.Set("ip_range", n.IPRange.String())
	d.Set("name", n.Name)
	d.Set("labels", n.Labels)
}


func waitForNetworkAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action, network *hcloud.Network) error {
	log.Printf("[INFO] Network (%d) waiting for %q action to complete...", network.ID, action.Command)
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] Network (%d) %q action succeeded", network.ID, action.Command)
	return nil
}
