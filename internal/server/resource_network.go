package server

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
	"github.com/hetznercloud/terraform-provider-hcloud/internal/merge"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/network"
)

// NetworkResourceType is the type name of the Hetzner Cloud Server
// network resource.
const NetworkResourceType = "hcloud_server_network"

// NetworkResource creates a Terraform schema for the hcloud_server_network
// resource.
func NetworkResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerNetworkCreate,
		ReadContext:   resourceServerNetworkRead,
		UpdateContext: resourceServerNetworkUpdate,
		DeleteContext: resourceServerNetworkDelete,
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
			"server_id": {
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
			"alias_ips": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServerNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
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

	server := &hcloud.Server{ID: d.Get("server_id").(int)}
	n := &hcloud.Network{ID: networkID.(int)}
	aliasIPs := make([]net.IP, 0, d.Get("alias_ips").(*schema.Set).Len())
	for _, aliasIP := range d.Get("alias_ips").(*schema.Set).List() {
		ip := net.ParseIP(aliasIP.(string))
		aliasIPs = append(aliasIPs, ip)
	}

	err := attachServerToNetwork(ctx, client, server, n, ip, aliasIPs)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	d.SetId(generateServerNetworkID(server, n))

	return resourceServerNetworkRead(ctx, d, m)
}

func resourceServerNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	server, network, _, err := lookupServerNetworkID(ctx, d.Id(), client)
	if err == errInvalidServerNetworkID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil {
		log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if network == nil {
		log.Printf("[WARN] Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if d.HasChange("alias_ips") {
		if err := updateServerAliasIPs(ctx, client, server, network, d.Get("alias_ips").(*schema.Set)); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	return resourceServerNetworkRead(ctx, d, m)
}

func resourceServerNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	server, network, privateNet, err := lookupServerNetworkID(ctx, d.Id(), client)
	if err == errInvalidServerNetworkID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if server == nil {
		log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if network == nil {
		log.Printf("[WARN] Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if privateNet == nil {
		log.Printf("[WARN] Server Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	d.SetId(generateServerNetworkID(server, network))
	setServerNetworkSchema(d, server, network, privateNet)
	return nil
}

func resourceServerNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	server, network, _, err := lookupServerNetworkID(ctx, d.Id(), client)
	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err := detachServerFromNetwork(ctx, client, server, network); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return nil
}

func setServerNetworkSchema(d *schema.ResourceData, server *hcloud.Server, network *hcloud.Network, serverPrivateNet *hcloud.ServerPrivateNet) {
	d.SetId(generateServerNetworkID(server, network))
	d.Set("ip", serverPrivateNet.IP.String())

	// We need to ensure that order of the list of alias_ips is kept stable no
	// matter what the Hetzner Cloud API returns. Therefore we merge the
	// returned IPs with the currently known alias_ips.
	tfAliasIPs := d.Get("alias_ips").(*schema.Set).List()
	aliasIPs := make([]string, len(tfAliasIPs))
	for i, v := range tfAliasIPs {
		aliasIPs[i] = v.(string)
	}
	hcAliasIPs := make([]string, len(serverPrivateNet.Aliases))
	for i, ip := range serverPrivateNet.Aliases {
		hcAliasIPs[i] = ip.String()
	}
	aliasIPs = merge.StringSlice(aliasIPs, hcAliasIPs)
	d.Set("alias_ips", aliasIPs)

	d.Set("mac_address", serverPrivateNet.MACAddress)
	if subnetID, ok := d.GetOk("subnet_id"); ok {
		d.Set("subnet_id", subnetID.(string))
	} else {
		d.Set("network_id", network.ID)
	}
	d.Set("server_id", server.ID)
}

func attachServerToNetwork(ctx context.Context, c *hcloud.Client, srv *hcloud.Server, nw *hcloud.Network, ip net.IP, aliasIPs []net.IP) error {
	var a *hcloud.Action

	opts := hcloud.ServerAttachToNetworkOpts{
		Network:  nw,
		IP:       ip,
		AliasIPs: aliasIPs,
	}

	err := control.Retry(control.DefaultRetries, func() error {
		var err error

		a, _, err = c.Server.AttachToNetwork(ctx, srv, opts)
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) ||
			hcloud.IsError(err, hcloud.ErrorCodeLocked) ||
			hcloud.IsError(err, hcloud.ErrorCodeServiceError) ||
			hcloud.IsError(err, hcloud.ErrorCodeNoSubnetAvailable) {
			return err
		}
		if err != nil {
			return control.AbortRetry(err)
		}
		return nil
	})
	if hcloud.IsError(err, hcloud.ErrorCodeServerAlreadyAttached) {
		log.Printf("[INFO] Server (%v) already attachted to network %v", srv.ID, nw.ID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("attach server to network: %v", err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return fmt.Errorf("attach server to network: %v", err)
	}
	return nil
}

func generateServerNetworkID(server *hcloud.Server, network *hcloud.Network) string {
	return fmt.Sprintf("%d-%d", server.ID, network.ID)
}

var errInvalidServerNetworkID = errors.New("invalid server network id")

// lookupServerNetworkID parses the terraform server network record id and return the server, network and the ServerPrivateNet
//
// id format: <server id>-<network id>
// Examples:
// 123-456
func lookupServerNetworkID(ctx context.Context, terraformID string, client *hcloud.Client) (server *hcloud.Server, network *hcloud.Network, serverPrivateNet *hcloud.ServerPrivateNet, err error) {
	if terraformID == "" {
		err = errInvalidServerNetworkID
		return
	}
	parts := strings.SplitN(terraformID, "-", 2)
	if len(parts) != 2 {
		err = errInvalidServerNetworkID
		return
	}

	serverID, err := strconv.Atoi(parts[0])
	if err != nil {
		err = errInvalidServerNetworkID
		return
	}

	server, _, err = client.Server.GetByID(ctx, serverID)
	if err != nil {
		err = errInvalidServerNetworkID
		return
	}
	if server == nil {
		err = errInvalidServerNetworkID
		return
	}

	networkID, err := strconv.Atoi(parts[1])
	if err != nil {
		err = errInvalidServerNetworkID
		return
	}

	network, _, err = client.Network.GetByID(ctx, networkID)
	if network == nil {
		err = errInvalidServerNetworkID
		return
	}

	for _, pn := range server.PrivateNet {
		if pn.Network.ID == network.ID {
			pn := pn
			serverPrivateNet = &pn
			return
		}
	}
	return
}

func updateServerAliasIPs(ctx context.Context, c *hcloud.Client, s *hcloud.Server, n *hcloud.Network, aliasIPs *schema.Set) error {
	const op = "hcloud/updateServerAliasIPs"

	opts := hcloud.ServerChangeAliasIPsOpts{
		Network:  n,
		AliasIPs: make([]net.IP, aliasIPs.Len()),
	}
	for i, v := range aliasIPs.List() {
		opts.AliasIPs[i] = net.ParseIP(v.(string))
	}
	a, _, err := c.Server.ChangeAliasIPs(ctx, s, opts)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

func detachServerFromNetwork(ctx context.Context, c *hcloud.Client, s *hcloud.Server, n *hcloud.Network) error {
	const op = "hcloud/detachServerFromNetwork"
	var a *hcloud.Action

	err := control.Retry(control.DefaultRetries, func() error {
		var err error

		a, _, err = c.Server.DetachFromNetwork(ctx, s, hcloud.ServerDetachFromNetworkOpts{Network: n})
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
	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}
