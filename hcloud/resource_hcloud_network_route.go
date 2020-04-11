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

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceNetworkRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkRouteCreate,
		Read:   resourceNetworkRouteRead,
		Delete: resourceNetworkRouteDelete,
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkRouteCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	_, destination, err := net.ParseCIDR(d.Get("destination").(string))
	if err != nil {
		return err
	}

	gateway := net.ParseIP(d.Get("gateway").(string))
	if gateway == nil {
		log.Printf("[WARN] Invalid gateway (%s), removing from state.", gateway)
		d.SetId("")
		return nil
	}
	networkID := d.Get("network_id")
	network := &hcloud.Network{ID: networkID.(int)}
	opts := hcloud.NetworkAddRouteOpts{
		Route: hcloud.NetworkRoute{
			Destination: destination,
			Gateway:     gateway,
		},
	}

	action, _, err := client.Network.AddRoute(ctx, network, opts)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			log.Printf("[INFO] Network (%v) conflict, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceNetworkRouteCreate(d, m)
		}
		return err
	}
	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}
	d.SetId(generateNetworkRouteID(network, destination.String()))

	return resourceNetworkRouteRead(d, m)
}

func resourceNetworkRouteRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	network, route, err := lookupNetworkRouteID(ctx, d.Id(), client)
	if err == errInvalidNetworkRouteID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}
	if network == nil {
		log.Printf("[WARN] Network Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	d.SetId(generateNetworkRouteID(network, route.Destination.String()))
	setNetworkRouteSchema(d, network, route)
	return nil

}

func resourceNetworkRouteDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	network, route, err := lookupNetworkRouteID(ctx, d.Id(), client)

	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	action, _, err := client.Network.DeleteRoute(ctx, network, hcloud.NetworkDeleteRouteOpts{
		Route: route,
	})
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// network route has already been deleted
			return nil
		} else if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			log.Printf("[INFO] Network (%v) conflict, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceNetworkRouteDelete(d, m)
		}
		return err
	}
	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}
	return nil
}

func setNetworkRouteSchema(d *schema.ResourceData, n *hcloud.Network, s hcloud.NetworkRoute) {
	d.SetId(generateNetworkRouteID(n, s.Destination.String()))
	d.Set("network_id", n.ID)
	d.Set("destination", s.Destination.String())
	d.Set("gateway", s.Gateway.String())
}

func generateNetworkRouteID(network *hcloud.Network, destination string) string {
	return fmt.Sprintf("%d-%s", network.ID, destination)
}

var errInvalidNetworkRouteID = errors.New("invalid network route id")

// lookupNetworkRouteID parses the terraform network route record id and return the network and route
//
// id format: <network id>-<ip range>
// Examples:
// 123-192.168.100.1/32 (network route of network 123 with the destination 192.168.100.1/32)
func lookupNetworkRouteID(ctx context.Context, terraformID string, client *hcloud.Client) (network *hcloud.Network, route hcloud.NetworkRoute, err error) {
	if terraformID == "" {
		err = errInvalidNetworkRouteID
		return
	}
	parts := strings.SplitN(terraformID, "-", 2)
	if len(parts) != 2 {
		err = errInvalidNetworkRouteID
		return
	}

	networkID, err := strconv.Atoi(parts[0])
	if err != nil {
		err = errInvalidNetworkRouteID
		return
	}

	_, destination, err := net.ParseCIDR(parts[1])
	if destination == nil || err != nil {
		err = errInvalidNetworkRouteID
		return
	}

	network, _, err = client.Network.GetByID(ctx, networkID)
	if network == nil {
		err = errInvalidNetworkRouteID
		return
	}

	for _, r := range network.Routes {
		if r.Destination.String() == destination.String() {
			route = r
			return
		}
	}
	err = errInvalidNetworkRouteID
	return
}
