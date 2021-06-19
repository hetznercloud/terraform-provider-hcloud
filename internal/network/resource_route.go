package network

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// RouteResourceType is the type name of the Hetzner Cloud Network Route resource.
const RouteResourceType = "hcloud_network_route"

// RouteResource creates a Terraform schema for the hcloud_network_route
// resource.
func RouteResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkRouteCreate,
		ReadContext:   resourceNetworkRouteRead,
		DeleteContext: resourceNetworkRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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

func resourceNetworkRouteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var a *hcloud.Action

	c := m.(*hcloud.Client)

	_, destination, err := net.ParseCIDR(d.Get("destination").(string))
	if err != nil {
		return hcclient.ErrorToDiag(err)
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

	err = control.Retry(control.DefaultRetries, func() error {
		var err error

		a, _, err = c.Network.AddRoute(ctx, network, opts)
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			return err
		}
		return control.AbortRetry(err)
	})
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	d.SetId(generateNetworkRouteID(network, destination.String()))

	return resourceNetworkRouteRead(ctx, d, m)
}

func resourceNetworkRouteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	network, route, err := lookupNetworkRouteID(ctx, d.Id(), client)
	if err == errInvalidNetworkRouteID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
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

func resourceNetworkRouteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var action *hcloud.Action

	c := m.(*hcloud.Client)

	network, route, err := lookupNetworkRouteID(ctx, d.Id(), c)
	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	err = control.Retry(control.DefaultRetries, func() error {
		var err error

		action, _, err = c.Network.DeleteRoute(ctx, network, hcloud.NetworkDeleteRouteOpts{
			Route: route,
		})
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			return err
		}
		return control.AbortRetry(err)
	})
	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		// network route has already been deleted
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
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
