package loadbalancer

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// TargetResourceType is the type name of the Hetzner Cloud Load Balancer
// target resource.
const TargetResourceType = "hcloud_load_balancer_target"

var errLoadBalancerTargetNotFound = errors.New("load balancer target not found")

// TargetResource creates a Terraform schema for the
// hcloud_load_balancer_target resource.
func TargetResource() *schema.Resource {
	targetProps := []string{"server_id", "label_selector", "ip"}
	return &schema.Resource{
		CreateContext: resourceLoadBalancerTargetCreate,
		ReadContext:   resourceLoadBalancerTargetRead,
		UpdateContext: resourceLoadBalancerTargetUpdate,
		DeleteContext: resourceLoadBalancerTargetDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					string(hcloud.LoadBalancerTargetTypeServer),
					string(hcloud.LoadBalancerTargetTypeLabelSelector),
					string(hcloud.LoadBalancerTargetTypeIP),
				}, false),
				Required: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"server_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: targetProps,
				ForceNew:     true,
			},
			"label_selector": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: targetProps,
				ForceNew:     true,
			},
			"ip": {
				Type:          schema.TypeString,
				Optional:      true,
				ExactlyOneOf:  targetProps,
				ConflictsWith: []string{"use_private_ip"},
				ForceNew:      true,
			},
			"use_private_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceLoadBalancerTargetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		lb  *hcloud.LoadBalancer
		tgt hcloud.LoadBalancerTarget
		a   *hcloud.Action
		err error
	)

	c := m.(*hcloud.Client)

	lbID := d.Get("load_balancer_id").(int)
	lb, _, err = c.LoadBalancer.GetByID(ctx, lbID)
	if err != nil {
		return diag.Errorf("get load balancer by id: %d: %v", lbID, err)
	}
	if lb == nil {
		return diag.Errorf("load balancer %d: not found", lbID)
	}

	tgtType := hcloud.LoadBalancerTargetType(d.Get("type").(string))
	switch tgtType {
	case hcloud.LoadBalancerTargetTypeServer:
		a, tgt, err = resourceLoadBalancerCreateServerTarget(ctx, c, lb, d)
	case hcloud.LoadBalancerTargetTypeLabelSelector:
		a, tgt, err = resourceLoadBalancerCreateLabelSelectorTarget(ctx, c, lb, d)
	case hcloud.LoadBalancerTargetTypeIP:
		a, tgt, err = resourceLoadBalancerCreateIPTarget(ctx, c, lb, d)
	default:
		return diag.Errorf("unsupported target type: %s", tgtType)
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if a != nil {
		if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
			return diag.Errorf("add load balancer target: %v", err)
		}
	}
	setLoadBalancerTarget(d, lbID, tgt)
	return nil
}

func resourceLoadBalancerCreateServerTarget(
	ctx context.Context, client *hcloud.Client, lb *hcloud.LoadBalancer, d *schema.ResourceData,
) (*hcloud.Action, hcloud.LoadBalancerTarget, error) {
	const op = "hcloud/resourceLoadBalancerCreateServerTarget"

	var (
		usePrivateIP bool
		tgt          hcloud.LoadBalancerTarget
	)

	sid, ok := d.GetOk("server_id")
	if !ok {
		return nil, tgt, fmt.Errorf("target type server: missing server_id")
	}
	serverID := sid.(int)

	server, _, err := client.Server.GetByID(ctx, serverID)
	if err != nil {
		return nil, tgt, fmt.Errorf("get server by id: %d: %v", serverID, err)
	}
	if server == nil {
		return nil, tgt, fmt.Errorf("server %d: not found", serverID)
	}

	opts := hcloud.LoadBalancerAddServerTargetOpts{
		Server: server,
	}
	if v, ok := d.GetOk("use_private_ip"); ok {
		usePrivateIP = v.(bool)
		opts.UsePrivateIP = hcloud.Bool(usePrivateIP)
	}

	err = control.Retry(control.DefaultRetries, func() error {
		var err error

		if usePrivateIP && len(lb.PrivateNet) == 0 {
			return errors.New("no private networks")
		}
		lb, _, err = client.LoadBalancer.GetByID(ctx, lb.ID)
		if err != nil {
			return control.AbortRetry(err)
		}
		if lb == nil {
			return control.AbortRetry(fmt.Errorf("load balancer %d: not found", lb.ID))
		}
		return nil
	})
	if err != nil {
		return nil, tgt, fmt.Errorf("%s: load balancer %d: %v", op, lb.ID, err)
	}

	tgt = hcloud.LoadBalancerTarget{
		Type:         hcloud.LoadBalancerTargetTypeServer,
		Server:       &hcloud.LoadBalancerTargetServer{Server: server},
		UsePrivateIP: usePrivateIP,
	}
	action, _, err := client.LoadBalancer.AddServerTarget(ctx, lb, opts)
	if err != nil {
		if hcloud.IsError(err, "target_already_defined") { // TODO: use const when hcloud go is released
			return nil, tgt, nil
		}
		return nil, tgt, fmt.Errorf("add server target: %v", err)
	}

	return action, tgt, nil
}

func resourceLoadBalancerCreateLabelSelectorTarget(
	ctx context.Context, client *hcloud.Client, lb *hcloud.LoadBalancer, d *schema.ResourceData,
) (*hcloud.Action, hcloud.LoadBalancerTarget, error) {
	var (
		opts hcloud.LoadBalancerAddLabelSelectorTargetOpts
		tgt  hcloud.LoadBalancerTarget
	)

	if v, ok := d.GetOk("label_selector"); ok {
		opts.Selector = v.(string)
	}
	if opts.Selector == "" {
		return nil, tgt, fmt.Errorf("label_selector is missing")
	}

	if v, ok := d.GetOk("use_private_ip"); ok {
		opts.UsePrivateIP = hcloud.Bool(v.(bool))
	}

	tgt = hcloud.LoadBalancerTarget{
		Type: hcloud.LoadBalancerTargetTypeLabelSelector,
		LabelSelector: &hcloud.LoadBalancerTargetLabelSelector{
			Selector: opts.Selector,
		},
		UsePrivateIP: opts.UsePrivateIP != nil && *opts.UsePrivateIP,
	}

	action, _, err := client.LoadBalancer.AddLabelSelectorTarget(ctx, lb, opts)
	if err != nil && hcloud.IsError(err, "target_already_defined") {
		return nil, tgt, nil
	}
	if err != nil {
		return nil, tgt, fmt.Errorf("add label selector target: %v", err)
	}
	return action, tgt, nil
}

func resourceLoadBalancerCreateIPTarget(
	ctx context.Context, client *hcloud.Client, lb *hcloud.LoadBalancer, d *schema.ResourceData,
) (*hcloud.Action, hcloud.LoadBalancerTarget, error) {
	var (
		opts hcloud.LoadBalancerAddIPTargetOpts
		tgt  hcloud.LoadBalancerTarget
	)

	if v, ok := d.GetOk("ip"); ok {
		opts.IP = net.ParseIP(v.(string))
	}
	if opts.IP == nil {
		return nil, tgt, fmt.Errorf("ip is missing or invalid")
	}

	tgt = hcloud.LoadBalancerTarget{
		Type: hcloud.LoadBalancerTargetTypeIP,
		IP:   &hcloud.LoadBalancerTargetIP{IP: opts.IP.String()},
	}

	action, _, err := client.LoadBalancer.AddIPTarget(ctx, lb, opts)
	if err != nil && hcloud.IsError(err, "target_already_defined") {
		return nil, tgt, nil
	}
	if err != nil {
		return nil, tgt, fmt.Errorf("add label selector target: %v", err)
	}
	return action, tgt, nil
}

func resourceLoadBalancerTargetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	lbID := d.Get("load_balancer_id").(int)
	tgtType := hcloud.LoadBalancerTargetType(d.Get("type").(string))

	_, tgt, err := findLoadBalancerTarget(ctx, client, lbID, tgtType, d)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	setLoadBalancerTarget(d, lbID, tgt)
	return nil
}

func resourceLoadBalancerTargetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	lbID := d.Get("load_balancer_id").(int)
	tgtType := hcloud.LoadBalancerTargetType(d.Get("type").(string))

	lb, tgt, err := findLoadBalancerTarget(ctx, client, lbID, tgtType, d)
	if errors.Is(err, errLoadBalancerTargetNotFound) {
		return resourceLoadBalancerTargetCreate(ctx, d, m)
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := removeLoadBalancerTarget(ctx, client, lb, tgt); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return resourceLoadBalancerTargetCreate(ctx, d, m)
}

func resourceLoadBalancerTargetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	tgtType := hcloud.LoadBalancerTargetType(d.Get("type").(string))
	lbID := d.Get("load_balancer_id").(int)

	lb, tgt, err := findLoadBalancerTarget(ctx, client, lbID, tgtType, d)
	if errors.Is(err, errLoadBalancerTargetNotFound) {
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := removeLoadBalancerTarget(ctx, client, lb, tgt); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return nil
}

func removeLoadBalancerTarget(ctx context.Context, c *hcloud.Client, lb *hcloud.LoadBalancer, tgt hcloud.LoadBalancerTarget) error {
	var (
		a     *hcloud.Action
		err   error
		hcErr hcloud.Error
	)

	err = control.Retry(control.DefaultRetries, func() error {
		switch tgt.Type {
		case hcloud.LoadBalancerTargetTypeServer:
			a, _, err = c.LoadBalancer.RemoveServerTarget(ctx, lb, tgt.Server.Server)
		case hcloud.LoadBalancerTargetTypeLabelSelector:
			a, _, err = c.LoadBalancer.RemoveLabelSelectorTarget(ctx, lb, tgt.LabelSelector.Selector)
		case hcloud.LoadBalancerTargetTypeIP:
			a, _, err = c.LoadBalancer.RemoveIPTarget(ctx, lb, net.ParseIP(tgt.IP.IP))
		default:
			return control.AbortRetry(fmt.Errorf("unsupported target type: %s", tgt.Type))
		}
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) || hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			return err
		}
		return control.AbortRetry(err)
	})
	if errors.As(err, &hcErr) && (hcErr.Code == "load_balancer_target_not_found" || strings.Contains(hcErr.Message, "target not found")) {
		// Target has been deleted already (e.g. by deleting the
		// network it was attached to)
		return nil
	}
	if err != nil {
		return fmt.Errorf("remove server target: %v", err)
	}

	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return fmt.Errorf("remove server target: wait for action: %v", err)
	}
	return nil
}

func findLoadBalancerTarget(
	ctx context.Context, client *hcloud.Client, lbID int, tgtType hcloud.LoadBalancerTargetType, d *schema.ResourceData,
) (*hcloud.LoadBalancer, hcloud.LoadBalancerTarget, error) {
	var (
		serverID      int
		labelSelector string
		ip            string
	)

	lb, _, err := client.LoadBalancer.GetByID(ctx, lbID)
	if err != nil {
		return nil, hcloud.LoadBalancerTarget{}, fmt.Errorf("get load balancer by id: %d: %v", lbID, err)
	}
	if lb == nil {
		return nil, hcloud.LoadBalancerTarget{}, fmt.Errorf("load balancer %d: not found", lbID)
	}
	if v, ok := d.GetOk("server_id"); ok {
		serverID = v.(int)
	}
	if v, ok := d.GetOk("label_selector"); ok {
		labelSelector = v.(string)
	}
	if v, ok := d.GetOk("ip"); ok {
		ip = v.(string)
	}

	for _, tgt := range lb.Targets {
		switch tgt.Type {
		case hcloud.LoadBalancerTargetTypeServer:
			if tgt.Server.Server.ID == serverID {
				return lb, tgt, nil
			}
		case hcloud.LoadBalancerTargetTypeLabelSelector:
			if tgt.LabelSelector.Selector == labelSelector {
				return lb, tgt, nil
			}
		case hcloud.LoadBalancerTargetTypeIP:
			if tgt.IP.IP == ip {
				return lb, tgt, nil
			}
		default:
			return nil, hcloud.LoadBalancerTarget{}, fmt.Errorf("unsupported target type: %s", tgtType)
		}
	}

	return nil, hcloud.LoadBalancerTarget{}, errLoadBalancerTargetNotFound
}

func setLoadBalancerTarget(d *schema.ResourceData, lbID int, tgt hcloud.LoadBalancerTarget) {
	d.Set("type", tgt.Type)
	d.Set("load_balancer_id", lbID)
	d.Set("use_private_ip", tgt.UsePrivateIP)

	switch tgt.Type {
	case hcloud.LoadBalancerTargetTypeServer:
		d.Set("server_id", tgt.Server.Server.ID)
		tgtID := generateLoadBalancerServerTargetID(tgt.Server.Server, lbID)
		d.SetId(tgtID)
	case hcloud.LoadBalancerTargetTypeLabelSelector:
		d.Set("label_selector", tgt.LabelSelector.Selector)
		tgtID := generateLoadBalancerLabelSelectorTargetID(tgt.LabelSelector.Selector, lbID)
		d.SetId(tgtID)
	case hcloud.LoadBalancerTargetTypeIP:
		d.Set("ip", tgt.IP.IP)
		tgtID := generateLoadBalancerIPTargetID(tgt.IP.IP, lbID)
		d.SetId(tgtID)
	}
}

func generateLoadBalancerServerTargetID(srv *hcloud.Server, lbID int) string {
	return fmt.Sprintf("lb-srv-tgt-%d-%d", srv.ID, lbID)
}

func generateLoadBalancerLabelSelectorTargetID(selector string, lbID int) string {
	h := sha256.Sum256([]byte(selector))
	return fmt.Sprintf("lb-label-selector-tgt-%x-%d", h, lbID)
}

func generateLoadBalancerIPTargetID(ip string, lbID int) string {
	h := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("lb-ip-tgt-%x-%d", h, lbID)
}
