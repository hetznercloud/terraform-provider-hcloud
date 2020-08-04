package hcloud

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadBalancerCreate,
		Read:   resourceLoadBalancerRead,
		Update: resourceLoadBalancerUpdate,
		Delete: resourceLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"load_balancer_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipv4": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"network_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"network_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"algorithm": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"round_robin",
								"least_connections",
							}, false),
						},
					},
				},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			"target": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"server",
							}, false),
						},
						"server_id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"use_private_ip": {
							Type:       schema.TypeBool,
							Optional:   true,
							Default:    false,
							Deprecated: "Does not work. Use the hcloud_load_balancer_target resource instead.",
						},
					},
				},
			},
		},
	}
}

func resourceLoadBalancerCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	opts := hcloud.LoadBalancerCreateOpts{
		Name:             d.Get("name").(string),
		LoadBalancerType: &hcloud.LoadBalancerType{Name: d.Get("load_balancer_type").(string)},
	}
	if algorithm, ok := d.GetOk("algorithm"); ok {
		tmpAlgorithm := parseTerraformAlgorithm(algorithm.([]interface{}))
		opts.Algorithm = &tmpAlgorithm
	}
	if location, ok := d.GetOk("location"); ok {
		opts.Location = &hcloud.Location{Name: location.(string)}
	}
	if networkZone, ok := d.GetOk("network_zone"); ok {
		opts.NetworkZone = hcloud.NetworkZone(networkZone.(string))
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}
	if targets, ok := d.GetOk("target"); ok {
		opts.Targets = parseTerraformTarget(targets.([]interface{}))
	}

	res, _, err := client.LoadBalancer.Create(ctx, opts)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(res.LoadBalancer.ID))
	if err := waitForLoadBalancerAction(ctx, client, res.Action, res.LoadBalancer); err != nil {
		return err
	}

	return resourceLoadBalancerRead(d, m)
}

func resourceLoadBalancerRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	loadBalancer, _, err := client.LoadBalancer.Get(ctx, d.Id())
	if err != nil {
		if resourceLoadBalancerIsNotFound(err, d) {
			return nil
		}
		return err
	}
	if loadBalancer == nil {
		d.SetId("")
		return nil
	}
	setLoadBalancerSchema(d, loadBalancer)
	return nil
}

func resourceLoadBalancerUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	loadBalancer, _, err := client.LoadBalancer.Get(ctx, d.Id())
	if err != nil {
		return err
	}
	if loadBalancer == nil {
		d.SetId("")
		return nil
	}

	d.Partial(true)
	if d.HasChange("name") {
		newName := d.Get("name")
		_, _, err := client.LoadBalancer.Update(ctx, loadBalancer, hcloud.LoadBalancerUpdateOpts{
			Name: newName.(string),
		})
		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return err
		}
		d.SetPartial("name")
	}

	if d.HasChange("load_balancer_type") {
		newType := d.Get("load_balancer_type")
		_, _, err := client.LoadBalancer.ChangeType(ctx, loadBalancer, hcloud.LoadBalancerChangeTypeOpts{
			LoadBalancerType: &hcloud.LoadBalancerType{Name: newType.(string)},
		})
		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return err
		}
		d.SetPartial("load_balancer_type")
	}

	if d.HasChange("algorithm") {
		algorithm := d.Get("algorithm")
		hcloudAlgorithm := parseTerraformAlgorithm(algorithm.([]interface{}))
		action, _, err := client.LoadBalancer.ChangeAlgorithm(ctx, loadBalancer, hcloud.LoadBalancerChangeAlgorithmOpts{
			Type: hcloudAlgorithm.Type,
		})

		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return err
		}
		if err := waitForLoadBalancerAction(ctx, client, action, loadBalancer); err != nil {
			return err
		}
		d.SetPartial("algorithm")
	}

	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.LoadBalancer.Update(ctx, loadBalancer, hcloud.LoadBalancerUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return err
		}
		d.SetPartial("labels")
	}

	if d.HasChange("target") {
		tfTargets := d.Get("target")
		hcloudTargetOpts := parseTerraformTarget(tfTargets.([]interface{}))

		// first we delete all targets that are not in the list
		for _, liveTarget := range loadBalancer.Targets {
			foundServer := false
			for _, hcloudTarget := range hcloudTargetOpts {
				if hcloudTarget.Type == hcloud.LoadBalancerTargetTypeServer {
					if liveTarget.Server.Server.ID == hcloudTarget.Server.Server.ID {
						foundServer = true
					}
				}
			}
			if !foundServer && liveTarget.Type == hcloud.LoadBalancerTargetTypeServer {
				action, _, err := client.LoadBalancer.RemoveServerTarget(ctx, loadBalancer, liveTarget.Server.Server)
				if err != nil {
					if resourceLoadBalancerIsNotFound(err, d) {
						return nil
					}
					return err
				}
				if err := waitForLoadBalancerAction(ctx, client, action, loadBalancer); err != nil {
					return err
				}
			}
		}

		// now we get the loadbalancer again
		loadBalancer, _, err := client.LoadBalancer.Get(ctx, d.Id())
		if err != nil {
			return err
		}
		if loadBalancer == nil {
			d.SetId("")
			return nil
		}
		// then we add all targets that are not in there already
		for _, hcloudTarget := range hcloudTargetOpts {
			foundServer := false
			for _, liveTarget := range loadBalancer.Targets {
				if hcloudTarget.Type == hcloud.LoadBalancerTargetTypeServer {
					if liveTarget.Server.Server.ID == hcloudTarget.Server.Server.ID {
						foundServer = true
					}
				}
			}
			if !foundServer && hcloudTarget.Type == hcloud.LoadBalancerTargetTypeServer {
				opts := hcloud.LoadBalancerAddServerTargetOpts{
					Server:       hcloudTarget.Server.Server,
					UsePrivateIP: hcloudTarget.UsePrivateIP,
				}
				action, _, err := client.LoadBalancer.AddServerTarget(ctx, loadBalancer, opts)
				if err != nil {
					if resourceLoadBalancerIsNotFound(err, d) {
						return nil
					}
					return err
				}
				if err := waitForLoadBalancerAction(ctx, client, action, loadBalancer); err != nil {
					return err
				}
			}
		}
		d.SetPartial("target")
	}
	d.Partial(false)
	return resourceLoadBalancerRead(d, m)
}

func resourceLoadBalancerDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	loadBalancer, _, err := client.LoadBalancer.Get(ctx, d.Id())
	if err != nil {
		return err
	}
	if loadBalancer == nil {
		d.SetId("")
		return nil
	}

	if _, err := client.LoadBalancer.Delete(ctx, loadBalancer); err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// loadBalancer has already been deleted
			return nil
		}
		return err
	}

	return nil
}

func resourceLoadBalancerIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] Load Balancer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setLoadBalancerSchema(d *schema.ResourceData, lb *hcloud.LoadBalancer) {
	d.SetId(strconv.Itoa(lb.ID))
	d.Set("id", strconv.Itoa(lb.ID))
	d.Set("name", lb.Name)
	d.Set("load_balancer_type", lb.LoadBalancerType.Name)
	d.Set("ipv4", lb.PublicNet.IPv4.IP.String())
	d.Set("ipv6", lb.PublicNet.IPv6.IP.String())
	d.Set("location", lb.Location.Name)
	d.Set("algorithm", algorithmToTerraformAlgorithm(lb.Algorithm))
	d.Set("network_zone", lb.Location.NetworkZone)
	d.Set("labels", lb.Labels)
	d.Set("targets", targetToTerraformTargets(lb.Targets))

	if len(lb.PrivateNet) > 0 {
		d.Set("network_id", lb.PrivateNet[0].Network.ID)
		d.Set("network_ip", lb.PrivateNet[0].IP.String())
	}
}

func waitForLoadBalancerAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action, loadBalancer *hcloud.LoadBalancer) error {
	log.Printf("[INFO] LoadBalancer (%d) waiting for %q action to complete...", loadBalancer.ID, action.Command)
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] LoadBalancer (%d) %q action succeeded", loadBalancer.ID, action.Command)
	return nil
}

func parseTerraformTarget(tfTargets []interface{}) (opts []hcloud.LoadBalancerCreateOptsTarget) {
	for _, _tfTarget := range tfTargets {
		tfTarget := _tfTarget.(map[string]interface{})
		opt := hcloud.LoadBalancerCreateOptsTarget{
			Type: hcloud.LoadBalancerTargetType(tfTarget["type"].(string)),
		}
		if serverID, ok := tfTarget["server_id"]; ok {
			opt.Server = hcloud.LoadBalancerCreateOptsTargetServer{Server: &hcloud.Server{ID: serverID.(int)}}
		}
		opts = append(opts, opt)
	}
	return
}

func targetToTerraformTargets(targets []hcloud.LoadBalancerTarget) (tfTargets []map[string]interface{}) {
	for _, target := range targets {
		tfTarget := make(map[string]interface{})
		tfTarget["type"] = string(target.Type)
		if target.Type == hcloud.LoadBalancerTargetTypeServer {
			tfTarget["server_id"] = target.Server.Server.ID
		}
		tfTargets = append(tfTargets, tfTarget)
	}
	return
}

func parseTerraformAlgorithm(tfAlgorithms []interface{}) (algorithm hcloud.LoadBalancerAlgorithm) {
	algorithm.Type = hcloud.LoadBalancerAlgorithmType(tfAlgorithms[0].(map[string]interface{})["type"].(string))
	return
}

func algorithmToTerraformAlgorithm(algorithm hcloud.LoadBalancerAlgorithm) (tfAlgorithms []map[string]interface{}) {
	tfAlgorithm := make(map[string]interface{})
	tfAlgorithm["type"] = string(algorithm.Type)
	tfAlgorithms = append(tfAlgorithms, tfAlgorithm)
	return
}
