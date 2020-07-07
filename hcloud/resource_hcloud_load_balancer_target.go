package hcloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

var errLoadBalancerTargetNotFound = errors.New("load balancer target not found")

func resourceLoadBalancerTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadBalancerTargetCreate,
		Read:   resourceLoadBalancerTargetRead,
		Update: resourceLoadBalancerTargetUpdate,
		Delete: resourceLoadBalancerTargetDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"server"}, false),
				Required:     true,
			},
			"load_balancer_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"use_private_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceLoadBalancerTargetCreate(d *schema.ResourceData, m interface{}) error {
	var usePrivateIP bool

	client := m.(*hcloud.Client)
	ctx := context.Background()

	tgtType := d.Get("type").(string)
	if tgtType != "server" {
		return fmt.Errorf("unsupported target type: %s", tgtType)
	}

	lbID := d.Get("load_balancer_id").(int)
	lb, _, err := client.LoadBalancer.GetByID(ctx, lbID)
	if err != nil {
		return fmt.Errorf("get load balancer by id: %d: %v", lbID, err)
	}
	if lb == nil {
		return fmt.Errorf("load balancer %d: not found", lbID)
	}

	sid, ok := d.GetOk("server_id")
	if !ok {
		return fmt.Errorf("target type server: missing server_id")
	}
	serverID := sid.(int)

	server, _, err := client.Server.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("get server by id: %d: %v", serverID, err)
	}
	if server == nil {
		return fmt.Errorf("server %d: not found", serverID)
	}

	opts := hcloud.LoadBalancerAddServerTargetOpts{
		Server: server,
	}
	if v, ok := d.GetOk("use_private_ip"); ok {
		usePrivateIP = v.(bool)
		opts.UsePrivateIP = hcloud.Bool(usePrivateIP)
	}

	if usePrivateIP && len(lb.PrivateNet) == 0 {
		log.Printf("[INFO] Load balancer (%d) not (yet) attached to a network. Retrying in one second", lb.ID)
		time.Sleep(time.Second)

		lb, _, err = client.LoadBalancer.GetByID(ctx, lbID)
		if err != nil {
			return fmt.Errorf("get load balancer by id: %d: %v", lbID, err)
		}
		if lb == nil {
			return fmt.Errorf("load balancer %d: not found", lbID)
		}
	}

	action, _, err := retryCodeConflict(func() (*hcloud.Action, *hcloud.Response, error) {
		return client.LoadBalancer.AddServerTarget(ctx, lb, opts)
	})
	if err != nil {
		return fmt.Errorf("add server target: %v", err)
	}
	if err := waitForLoadBalancerAction(ctx, client, action, lb); err != nil {
		return fmt.Errorf("add server target: %v", err)
	}
	setLoadBalancerTarget(d, lb.ID, hcloud.LoadBalancerTarget{
		Type:         hcloud.LoadBalancerTargetTypeServer,
		Server:       &hcloud.LoadBalancerTargetServer{Server: server},
		UsePrivateIP: usePrivateIP,
	})
	return nil
}

func resourceLoadBalancerTargetRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	lbID := d.Get("load_balancer_id").(int)

	_, tgt, err := findLoadBalancerTarget(ctx, client, lbID, d)
	if err != nil {
		return err
	}

	setLoadBalancerTarget(d, lbID, tgt)
	return nil
}

func resourceLoadBalancerTargetUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	lbID := d.Get("load_balancer_id").(int)

	lb, tgt, err := findLoadBalancerTarget(ctx, client, lbID, d)
	if errors.Is(err, errLoadBalancerTargetNotFound) {
		return resourceLoadBalancerTargetCreate(d, m)
	}
	if err != nil {
		return err
	}

	action, _, err := retryCodeConflict(func() (*hcloud.Action, *hcloud.Response, error) {
		return client.LoadBalancer.RemoveServerTarget(ctx, lb, tgt.Server.Server)
	})
	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		return resourceLoadBalancerTargetCreate(d, m)
	}
	if err != nil {
		return fmt.Errorf("remove existing target: %v", err)
	}
	if err := waitForLoadBalancerAction(ctx, client, action, lb); err != nil {
		return fmt.Errorf("remove existing target: %v", err)
	}

	return resourceLoadBalancerTargetCreate(d, m)
}

func resourceLoadBalancerTargetDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	lbID := d.Get("load_balancer_id").(int)

	lb, tgt, err := findLoadBalancerTarget(ctx, client, lbID, d)
	if errors.Is(err, errLoadBalancerTargetNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	action, _, err := retryCodeConflict(func() (*hcloud.Action, *hcloud.Response, error) {
		return client.LoadBalancer.RemoveServerTarget(ctx, lb, tgt.Server.Server)
	})
	if err != nil {
		if hcErr, ok := err.(hcloud.Error); ok {
			if hcErr.Code == "load_balancer_target_not_found" || strings.Contains(hcErr.Message, "target not found") {
				// Target has been deleted already (e.g. by deleting the
				// network it was attached to)
				return nil
			}
		}
		return fmt.Errorf("remove server target: %v", err)
	}
	if err := waitForLoadBalancerAction(ctx, client, action, lb); err != nil {
		return fmt.Errorf("remove server target: wait for action: %v", err)
	}

	return nil
}

func findLoadBalancerTarget(
	ctx context.Context, client *hcloud.Client, lbID int, d *schema.ResourceData,
) (*hcloud.LoadBalancer, hcloud.LoadBalancerTarget, error) {
	var serverID int

	lb, _, err := client.LoadBalancer.GetByID(ctx, lbID)
	if err != nil {
		return nil, hcloud.LoadBalancerTarget{}, fmt.Errorf("get load balancer by id: %d: %v", lbID, err)
	}
	if lb == nil {
		return nil, hcloud.LoadBalancerTarget{}, fmt.Errorf("load balancer %d: not found", lbID)
	}
	if sid, ok := d.GetOk("server_id"); ok {
		serverID = sid.(int)
	}

	for _, tgt := range lb.Targets {
		if tgt.Type == hcloud.LoadBalancerTargetTypeServer && tgt.Server.Server.ID == serverID {
			return lb, tgt, nil
		}
	}
	return nil, hcloud.LoadBalancerTarget{}, errLoadBalancerTargetNotFound
}

func setLoadBalancerTarget(d *schema.ResourceData, lbID int, tgt hcloud.LoadBalancerTarget) {
	d.Set("type", tgt.Type)
	d.Set("load_balancer_id", lbID)
	d.Set("use_private_ip", tgt.UsePrivateIP)

	if tgt.Type == hcloud.LoadBalancerTargetTypeServer {
		d.Set("server_id", tgt.Server.Server.ID)

		tgtID := generateLoadBalancerServerTargetID(tgt.Server.Server, lbID)
		d.SetId(tgtID)
	}
}

func generateLoadBalancerServerTargetID(srv *hcloud.Server, lbID int) string {
	return fmt.Sprintf("lb-srv-tgt-%d-%d", srv.ID, lbID)
}
