package loadbalancer

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
)

// NetworkResourceType is the type name of the Hetzner Cloud Load Balancer
// network resource.
const NetworkResourceType = "hcloud_load_balancer_network"

// NetworkResource creates a Terraform schema for the
// hcloud_load_balancer_network resource.
func NetworkResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLoadBalancerNetworkCreate,
		ReadContext:   resourceLoadBalancerNetworkRead,
		DeleteContext: resourceLoadBalancerNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"enable_public_interface": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
		},
	}
}

func resourceLoadBalancerNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var action *hcloud.Action

	c := m.(*hcloud.Client)

	ip := net.ParseIP(d.Get("ip").(string))

	networkID, nwIDSet := d.GetOk("network_id")
	subNetID, snIDSet := d.GetOk("subnet_id")
	if (nwIDSet && snIDSet) || (!nwIDSet && !snIDSet) {
		return diag.Errorf("either network_id or subnet_id must be set")
	}

	if snIDSet {
		nwID, _, err := network.ParseSubnetID(subNetID.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		networkID = nwID
	}

	loadBalancerID := d.Get("load_balancer_id").(int)
	lb := &hcloud.LoadBalancer{ID: loadBalancerID}

	nw := &hcloud.Network{ID: networkID.(int)}
	opts := hcloud.LoadBalancerAttachToNetworkOpts{
		Network: nw,
		IP:      ip,
	}

	err := control.Retry(control.DefaultRetries, func() error {
		var err error

		action, _, err = c.LoadBalancer.AttachToNetwork(ctx, lb, opts)
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) ||
			hcloud.IsError(err, hcloud.ErrorCodeLocked) ||
			hcloud.IsError(err, hcloud.ErrorCodeServiceError) ||
			hcloud.IsError(err, hcloud.ErrorCodeNoSubnetAvailable) {
			// Retry on any of the above listed errors
			return err
		}

		return control.AbortRetry(err)
	})
	if hcloud.IsError(err, hcloud.ErrorCodeLoadBalancerAlreadyAttached) &&
		isLoadBalancerAttachedToNetwork(ctx, c, lb, nw) {
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	enablePublicInterface := d.Get("enable_public_interface").(bool)
	err = resourceLoadBalancerNetworkUpdatePublicInterface(ctx, enablePublicInterface, lb, c)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	d.SetId(generateLoadBalancerNetworkID(lb, nw))

	return resourceLoadBalancerNetworkRead(ctx, d, m)
}

func resourceLoadBalancerNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	server, network, privateNet, err := lookupLoadBalancerNetworkID(ctx, d.Id(), client)
	if err == errInvalidLoadBalancerNetworkID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil {
		log.Printf("[WARN] LoadBalancer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if network == nil {
		log.Printf("[WARN] Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if privateNet == nil {
		log.Printf("[WARN] LoadBalancer Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	d.SetId(generateLoadBalancerNetworkID(server, network))
	setLoadBalancerNetworkSchema(d, server, network, privateNet)
	return nil
}

func resourceLoadBalancerNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var action *hcloud.Action

	client := m.(*hcloud.Client)

	server, network, _, err := lookupLoadBalancerNetworkID(ctx, d.Id(), client)

	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}

	err = control.Retry(control.DefaultRetries, func() error {
		var err error

		action, _, err = client.LoadBalancer.DetachFromNetwork(ctx, server, hcloud.LoadBalancerDetachFromNetworkOpts{
			Network: network,
		})
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) ||
			hcloud.IsError(err, hcloud.ErrorCodeLocked) ||
			hcloud.IsError(err, hcloud.ErrorCodeServiceError) {
			return err
		}
		return control.AbortRetry(err)
	})

	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		// network has already been deleted
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &client.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func setLoadBalancerNetworkSchema(d *schema.ResourceData, loadBalancer *hcloud.LoadBalancer, network *hcloud.Network, loadBalancerPrivateNet *hcloud.LoadBalancerPrivateNet) {
	d.SetId(generateLoadBalancerNetworkID(loadBalancer, network))
	d.Set("ip", loadBalancerPrivateNet.IP.String())
	d.Set("enable_public_interface", loadBalancer.PublicNet.Enabled)
	d.Set("load_balancer_id", loadBalancer.ID)
	if subnetID, ok := d.GetOk("subnet_id"); ok {
		d.Set("subnet_id", subnetID)
	} else {
		d.Set("network_id", network.ID)
	}
}

func isLoadBalancerAttachedToNetwork(
	ctx context.Context, c *hcloud.Client, lb *hcloud.LoadBalancer, n *hcloud.Network,
) bool {
	lbID := lb.ID
	lb, _, err := c.LoadBalancer.GetByID(ctx, lbID)
	if lb == nil || err != nil {
		log.Printf("[WARN] Failed to retrieve load balancer with id %d", lbID)
		return false
	}
	for _, privNet := range lb.PrivateNet {
		if privNet.Network != nil && privNet.Network.ID == n.ID {
			return true
		}
	}
	return false
}

func generateLoadBalancerNetworkID(server *hcloud.LoadBalancer, network *hcloud.Network) string {
	return fmt.Sprintf("%d-%d", server.ID, network.ID)
}

var errInvalidLoadBalancerNetworkID = errors.New("invalid load balancer network id")

// lookupLoadBalancerNetworkID parses the terraform load balancer network record id and return the load balancer, network and the LoadBalancerPrivateNet
//
// id format: <load balancer id>-<network id>
// Examples:
// 123-456
func lookupLoadBalancerNetworkID(
	ctx context.Context, tfID string, c *hcloud.Client,
) (*hcloud.LoadBalancer, *hcloud.Network, *hcloud.LoadBalancerPrivateNet, error) {
	if tfID == "" {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}
	parts := strings.SplitN(tfID, "-", 2)
	if len(parts) != 2 {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}

	loadBalancerID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}

	loadBalancer, _, err := c.LoadBalancer.GetByID(ctx, loadBalancerID)
	if err != nil {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}
	if loadBalancer == nil {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}

	networkID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}

	network, _, err := c.Network.GetByID(ctx, networkID)
	if err != nil {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}
	if network == nil {
		return nil, nil, nil, errInvalidLoadBalancerNetworkID
	}

	for _, pn := range loadBalancer.PrivateNet {
		if pn.Network.ID == network.ID {
			return loadBalancer, network, &pn, nil
		}
	}
	return nil, nil, nil, errInvalidLoadBalancerNetworkID
}

func resourceLoadBalancerNetworkUpdatePublicInterface(ctx context.Context, enable bool, lb *hcloud.LoadBalancer, client *hcloud.Client) error {
	var (
		action *hcloud.Action
		err    error
	)

	if enable {
		action, _, err = client.LoadBalancer.EnablePublicInterface(ctx, lb)
	} else {
		action, _, err = client.LoadBalancer.DisablePublicInterface(ctx, lb)
	}
	if err != nil {
		return err
	}
	return hcclient.WaitForAction(ctx, &client.Action, action)
}
