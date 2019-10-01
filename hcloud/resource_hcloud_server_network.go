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

func resourceServerNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerNetworkCreate,
		Read:   resourceServerNetworkRead,
		Update: resourceServerNetworkUpdate,
		Delete: resourceServerNetworkDelete,
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeInt,
				Required: true,
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
				Type:     schema.TypeList,
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

func resourceServerNetworkCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	ip := net.ParseIP(d.Get("ip").(string))
	networkID := d.Get("network_id")

	serverID := d.Get("server_id")

	server := &hcloud.Server{ID: serverID.(int)}

	network := &hcloud.Network{ID: networkID.(int)}
	opts := hcloud.ServerAttachToNetworkOpts{
		Network: network,
		IP:      ip,
	}

	action, _, err := client.Server.AttachToNetwork(ctx, server, opts)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			log.Printf("[INFO] Network (%v) conflict, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceServerNetworkCreate(d, m)
		} else if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			log.Printf("[INFO] Network (%v) locked, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceServerNetworkCreate(d, m)
		} else if hcloud.IsError(err, hcloud.ErrorCodeServerAlreadyAttached) {
			log.Printf("[INFO] Server (%v) already attachted to network %v", server.ID, network.ID)
			d.SetId(generateServerNetworkID(server, network))

			return resourceServerNetworkRead(d, m)
		}
		return err
	}
	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}
	d.SetId(generateServerNetworkID(server, network))

	return resourceServerNetworkRead(d, m)
}

func resourceServerNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	server, network, _, err := lookupServerNetworkID(ctx, d.Id(), client)
	if err == errInvalidServerNetworkID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
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
		opts := hcloud.ServerChangeAliasIPsOpts{
			Network: network,
		}
		for _, aliasIP := range d.Get("alias_ips").([]interface{}) {
			ip := net.ParseIP(aliasIP.(string))
			opts.AliasIPs = append(opts.AliasIPs, ip)
		}
		action, _, err := client.Server.ChangeAliasIPs(ctx, server, opts)
		if err != nil {
			return err
		}
		if err := waitForNetworkAction(ctx, client, action, network); err != nil {
			return err
		}
	}
	return resourceServerNetworkRead(d, m)
}

func resourceServerNetworkRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, network, privateNet, err := lookupServerNetworkID(ctx, d.Id(), client)
	if err == errInvalidServerNetworkID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
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

func resourceServerNetworkDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, network, _, err := lookupServerNetworkID(ctx, d.Id(), client)

	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	action, _, err := client.Server.DetachFromNetwork(ctx, server, hcloud.ServerDetachFromNetworkOpts{
		Network: network,
	})
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// network route has already been deleted
			return nil
		} else if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			log.Printf("[INFO] Network (%v) conflict, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceServerNetworkDelete(d, m)
		} else if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			log.Printf("[INFO] Network (%v) locked, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceServerNetworkDelete(d, m)
		}

		return err
	}
	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}
	return nil
}

func setServerNetworkSchema(d *schema.ResourceData, server *hcloud.Server, network *hcloud.Network, serverPrivateNet *hcloud.ServerPrivateNet) {
	d.SetId(generateServerNetworkID(server, network))
	d.Set("ip", serverPrivateNet.IP.String())
	var ips []string
	for _, ip := range serverPrivateNet.Aliases {
		ips = append(ips, ip.String())
	}
	d.Set("alias_ips", ips)
	d.Set("mac_address", serverPrivateNet.MACAddress)
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
			serverPrivateNet = &pn
			return
		}
	}
	return
}
