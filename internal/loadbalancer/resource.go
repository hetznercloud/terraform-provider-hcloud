package loadbalancer

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud Load Balancer resource.
const ResourceType = "hcloud_load_balancer"

// Resource creates a Terraform schema for the hcloud_load_balancer resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLoadBalancerCreate,
		ReadContext:   resourceLoadBalancerRead,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				Type:     schema.TypeSet,
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

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

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
		opts.Targets = parseTerraformTarget(targets.(*schema.Set))
	}

	res, _, err := c.LoadBalancer.Create(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	d.SetId(strconv.Itoa(res.LoadBalancer.ID))
	if err := hcclient.WaitForAction(ctx, &c.Action, res.Action); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return resourceLoadBalancerRead(ctx, d, m)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	loadBalancer, _, err := client.LoadBalancer.Get(ctx, d.Id())
	if err != nil {
		if resourceLoadBalancerIsNotFound(err, d) {
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}
	if loadBalancer == nil {
		d.SetId("")
		return nil
	}
	setLoadBalancerSchema(d, loadBalancer)
	return nil
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)
	loadBalancer, _, err := c.LoadBalancer.Get(ctx, d.Id())
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if loadBalancer == nil {
		d.SetId("")
		return nil
	}

	d.Partial(true)
	if d.HasChange("name") {
		newName := d.Get("name")
		_, _, err := c.LoadBalancer.Update(ctx, loadBalancer, hcloud.LoadBalancerUpdateOpts{
			Name: newName.(string),
		})
		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("load_balancer_type") {
		newType := d.Get("load_balancer_type")
		action, _, err := c.LoadBalancer.ChangeType(ctx, loadBalancer, hcloud.LoadBalancerChangeTypeOpts{
			LoadBalancerType: &hcloud.LoadBalancerType{Name: newType.(string)},
		})
		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
		if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("algorithm") {
		algorithm := d.Get("algorithm")
		hcloudAlgorithm := parseTerraformAlgorithm(algorithm.([]interface{}))
		ao := hcloud.LoadBalancerChangeAlgorithmOpts{ //nolint:gosimple
			Type: hcloudAlgorithm.Type,
		}
		action, _, err := c.LoadBalancer.ChangeAlgorithm(ctx, loadBalancer, ao)

		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
		if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := c.LoadBalancer.Update(ctx, loadBalancer, hcloud.LoadBalancerUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceLoadBalancerIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("target") {
		tfTargets := d.Get("target")
		hcloudTargetOpts := parseTerraformTarget(tfTargets.(*schema.Set))

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
				action, _, err := c.LoadBalancer.RemoveServerTarget(ctx, loadBalancer, liveTarget.Server.Server)
				if err != nil {
					if resourceLoadBalancerIsNotFound(err, d) {
						return nil
					}
					return hcclient.ErrorToDiag(err)
				}
				if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
					return hcclient.ErrorToDiag(err)
				}
			}
		}

		// now we get the loadbalancer again
		loadBalancer, _, err := c.LoadBalancer.Get(ctx, d.Id())
		if err != nil {
			return hcclient.ErrorToDiag(err)
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
				action, _, err := c.LoadBalancer.AddServerTarget(ctx, loadBalancer, opts)
				if err != nil {
					if resourceLoadBalancerIsNotFound(err, d) {
						return nil
					}
					return hcclient.ErrorToDiag(err)
				}
				if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
					return hcclient.ErrorToDiag(err)
				}
			}
		}
	}
	d.Partial(false)
	return resourceLoadBalancerRead(ctx, d, m)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	loadBalancer, _, err := client.LoadBalancer.Get(ctx, d.Id())
	if err != nil {
		return hcclient.ErrorToDiag(err)
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
		return hcclient.ErrorToDiag(err)
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
	d.Set("name", lb.Name)
	d.Set("load_balancer_type", lb.LoadBalancerType.Name)
	d.Set("ipv4", lb.PublicNet.IPv4.IP.String())
	d.Set("ipv6", lb.PublicNet.IPv6.IP.String())
	d.Set("location", lb.Location.Name)
	d.Set("algorithm", algorithmToTerraformAlgorithm(lb.Algorithm))
	d.Set("network_zone", lb.Location.NetworkZone)
	d.Set("labels", lb.Labels)
	d.Set("target", targetToTerraformTargets(lb.Targets))

	if len(lb.PrivateNet) > 0 {
		d.Set("network_id", lb.PrivateNet[0].Network.ID)
		d.Set("network_ip", lb.PrivateNet[0].IP.String())
	}
}

func parseTerraformTarget(tfTargets *schema.Set) (opts []hcloud.LoadBalancerCreateOptsTarget) {
	for _, _tfTarget := range tfTargets.List() {
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

func targetToTerraformTargets(targets []hcloud.LoadBalancerTarget) []map[string]interface{} {
	tfTargets := make([]map[string]interface{}, len(targets))
	for i, target := range targets {
		tfTarget := make(map[string]interface{})
		tfTarget["type"] = string(target.Type)
		if target.Type == hcloud.LoadBalancerTargetTypeServer {
			tfTarget["server_id"] = target.Server.Server.ID
		}
		tfTargets[i] = tfTarget
	}

	return tfTargets
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
