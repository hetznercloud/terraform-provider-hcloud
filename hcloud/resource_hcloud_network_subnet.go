package hcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceNetworkSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkSubnetCreate,
		Read:   resourceNetworkSubnetRead,
		Delete: resourceNetworkSubnetDelete,
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_range": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vswitch_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkSubnetCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	_, ipRange, err := net.ParseCIDR(d.Get("ip_range").(string))
	if err != nil {
		return err
	}
	networkID := d.Get("network_id")
	network := &hcloud.Network{ID: networkID.(int)}
	opts := hcloud.NetworkAddSubnetOpts{
		Subnet: &hcloud.NetworkSubnet{
			IPRange:     ipRange,
			NetworkZone: hcloud.NetworkZone(d.Get("network_zone").(string)),
			Type:        hcloud.NetworkSubnetType(d.Get("type").(string)),
			VSwitchID:   d.Get("vswitch_id").(int),
		},
	}

	action, _, err := client.Network.AddSubnet(ctx, network, opts)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			log.Printf("[INFO] Network (%v) conflict, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceNetworkSubnetCreate(d, m)
		}
		return err
	}

	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}
	d.SetId(generateNetworkSubnetID(network, ipRange.String()))

	return resourceNetworkSubnetRead(d, m)
}

func resourceNetworkSubnetRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	network, subnet, err := lookupNetworkSubnetID(ctx, d.Id(), client)
	if err == errInvalidNetworkSubnetID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}
	if network == nil {
		log.Printf("[WARN] Network Subnet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	d.SetId(generateNetworkSubnetID(network, subnet.IPRange.String()))
	setNetworkSubnetSchema(d, network, subnet)
	return nil

}

func resourceNetworkSubnetDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	network, subnet, err := lookupNetworkSubnetID(ctx, d.Id(), client)

	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	action, _, err := client.Network.RemoveSubnet(ctx, network, hcloud.NetworkRemoveSubnetOpts{
		Subnet: subnet,
	})
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// network subnet has already been deleted
			return nil
		} else if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			log.Printf("[INFO] Network (%v) conflict, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceNetworkSubnetDelete(d, m)
		}
		return err
	}
	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}
	return nil
}

func setNetworkSubnetSchema(d *schema.ResourceData, n *hcloud.Network, s *hcloud.NetworkSubnet) {
	d.SetId(generateNetworkSubnetID(n, s.IPRange.String()))
	d.Set("network_id", n.ID)
	d.Set("network_zone", s.NetworkZone)
	d.Set("ip_range", s.IPRange.String())
	d.Set("vswitch_id", s.VSwitchID)
	d.Set("type", s.Type)
	d.Set("gateway", s.Gateway.String())
}

func generateNetworkSubnetID(network *hcloud.Network, ipRange string) string {
	return fmt.Sprintf("%d-%s", network.ID, ipRange)
}

var errInvalidNetworkSubnetID = errors.New("invalid network subnet id")

// lookupNetworkSubnetID parses the terraform network subnet record id and return the network and subnet
//
// id format: <network id>-<ip range>
// Examples:
// 123-192.168.100.1/32 (network subnet of network 123 with the ip range 192.168.100.1/32)
func lookupNetworkSubnetID(ctx context.Context, terraformID string, client *hcloud.Client) (network *hcloud.Network, subnet *hcloud.NetworkSubnet, err error) {
	if terraformID == "" {
		err = errInvalidNetworkSubnetID
		return
	}
	parts := strings.SplitN(terraformID, "-", 2)
	if len(parts) != 2 {
		err = errInvalidNetworkSubnetID
		return
	}

	networkID, err := strconv.Atoi(parts[0])
	if err != nil {
		err = errInvalidNetworkSubnetID
		return
	}

	_, ipRange, err := net.ParseCIDR(parts[1])
	if ipRange == nil || err != nil {
		err = errInvalidNetworkSubnetID
		return
	}

	network, _, err = client.Network.GetByID(ctx, networkID)
	if network == nil {
		err = errInvalidNetworkSubnetID
		return
	}
	for _, sn := range network.Subnets {
		if sn.IPRange.String() == ipRange.String() {
			subnet = &sn
			return
		}
	}
	return
}
