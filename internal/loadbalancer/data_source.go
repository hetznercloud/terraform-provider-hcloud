package loadbalancer

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Load Balancer resource.
	DataSourceType = "hcloud_load_balancer"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Load Balancer resources.
	DataSourceListType = "hcloud_load_balancers"
)

// getCommonDataSchema returns a new common schema used by all load balancer data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"load_balancer_type": {
			Type:     schema.TypeString,
			Computed: true,
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
			ForceNew: true,
			Computed: true,
		},
		"network_zone": {
			Type:     schema.TypeString,
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
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"labels": {
			Type:     schema.TypeMap,
			Computed: true,
		},
		"target": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"server_id": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"label_selector": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"service": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"protocol": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"listen_port": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"destination_port": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"proxyprotocol": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"http": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"sticky_sessions": {
									Type:     schema.TypeBool,
									Computed: true,
								},
								"cookie_name": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"cookie_lifetime": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"certificates": {
									Type:     schema.TypeList,
									Computed: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"redirect_http": {
									Type:     schema.TypeBool,
									Computed: true,
								},
							},
						},
					},
					"health_check": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"protocol": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"port": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"interval": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"timeout": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"retries": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"http": {
									Type:     schema.TypeList,
									Computed: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"domain": {
												Type:     schema.TypeString,
												Computed: true,
											},
											"path": {
												Type:     schema.TypeString,
												Computed: true,
											},
											"response": {
												Type:     schema.TypeString,
												Computed: true,
											},
											"tls": {
												Type:     schema.TypeBool,
												Computed: true,
											},
											"status_codes": {
												Type:     schema.TypeList,
												Computed: true,
												Elem: &schema.Schema{
													Type: schema.TypeInt,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"delete_protection": {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}
}

// DataSource creates a new Terraform schema for the hcloud_load_balancer
// resource.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLoadBalancerRead,
		Schema: datasourceutil.MergeSchema(
			getCommonDataSchema(),
			map[string]*schema.Schema{
				"with_selector": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		),
	}
}

func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLoadBalancerListRead,
		Schema: map[string]*schema.Schema{
			"load_balancers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
			"with_selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHcloudLoadBalancerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	if id, ok := d.GetOk("id"); ok {
		lb, _, err := client.LoadBalancer.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if lb == nil {
			return diag.Errorf("no Load Balancer found with id %d", id)
		}
		setLoadBalancerSchema(d, lb)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		lb, _, err := client.LoadBalancer.GetByName(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if lb == nil {
			return diag.Errorf("no Load Balancer found with name %s", name)
		}
		setLoadBalancerSchema(d, lb)
		return nil
	}

	selector := d.Get("with_selector").(string)
	if selector != "" {
		var allLoadBalancers []*hcloud.LoadBalancer

		opts := hcloud.LoadBalancerListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allLoadBalancers, err := client.LoadBalancer.AllWithOpts(ctx, opts)
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if len(allLoadBalancers) == 0 {
			return diag.Errorf("no Load Balancer found for selector %q", selector)
		}
		if len(allLoadBalancers) > 1 {
			return diag.Errorf("more than one Load Balancer found for selector %q", selector)
		}
		setLoadBalancerSchema(d, allLoadBalancers[0])
		return nil
	}
	return diag.Errorf("please specify an id, a name or a selector to lookup the Load Balancer")
}

func dataSourceHcloudLoadBalancerListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector").(string)

	opts := hcloud.LoadBalancerListOpts{ListOpts: hcloud.ListOpts{LabelSelector: selector}}
	allLoadBalancers, err := client.LoadBalancer.AllWithOpts(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	ids := make([]string, len(allLoadBalancers))
	tfLoadBalancers := make([]map[string]interface{}, len(allLoadBalancers))
	for i, loadBalancer := range allLoadBalancers {
		ids[i] = strconv.Itoa(loadBalancer.ID)
		tfLoadBalancers[i] = getLoadBalancerAttributes(loadBalancer)
	}
	d.Set("load_balancers", tfLoadBalancers)
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))

	return nil
}
