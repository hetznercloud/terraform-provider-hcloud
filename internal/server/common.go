package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
)

func attachServerToNetwork(ctx context.Context, c *hcloud.Client, srv *hcloud.Server, nw *hcloud.Network, ip net.IP, aliasIPs []net.IP) error {
	var action *hcloud.Action

	opts := hcloud.ServerAttachToNetworkOpts{
		Network:  nw,
		IP:       ip,
		AliasIPs: aliasIPs,
	}

	err := control.Retry(control.DefaultRetries, func() error {
		var err error

		action, _, err = c.Server.AttachToNetwork(ctx, srv, opts)
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
		return fmt.Errorf("attach server to network: %w", err)
	}
	if err = c.Action.WaitFor(ctx, action); err != nil {
		return fmt.Errorf("attach server to network: %w", err)
	}
	return nil
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
	action, _, err := c.Server.ChangeAliasIPs(ctx, s, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err = c.Action.WaitFor(ctx, action); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func detachServerFromNetwork(ctx context.Context, c *hcloud.Client, s *hcloud.Server, n *hcloud.Network) error {
	const op = "hcloud/detachServerFromNetwork"
	var action *hcloud.Action

	err := control.Retry(control.DefaultRetries, func() error {
		var err error

		action, _, err = c.Server.DetachFromNetwork(ctx, s, hcloud.ServerDetachFromNetworkOpts{Network: n})
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) ||
			hcloud.IsError(err, hcloud.ErrorCodeLocked) ||
			hcloud.IsError(err, hcloud.ErrorCodeServiceError) {
			return err
		}
		return control.AbortRetry(err)
	})
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// network has already been deleted
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = c.Action.WaitFor(ctx, action); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
