package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudLoadBalancerRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"with_selector": {
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
		},
	}
}
func dataSourceHcloudLoadBalancerRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	var lb *hcloud.LoadBalancer
	if id, ok := d.GetOk("id"); ok {
		lb, _, err = client.LoadBalancer.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if lb == nil {
			return fmt.Errorf("no Load Balancer found with id %d", id)
		}
		setLoadBalancerSchema(d, lb)
		return
	}
	if name, ok := d.GetOk("name"); ok {
		lb, _, err = client.LoadBalancer.GetByName(ctx, name.(string))
		if err != nil {
			return err
		}
		if lb == nil {
			return fmt.Errorf("no Load Balancer found with name %s", name)
		}
		setLoadBalancerSchema(d, lb)
		return
	}

	selector := d.Get("with_selector").(string)
	if selector != "" {
		var allLoadBalancers []*hcloud.LoadBalancer

		opts := hcloud.LoadBalancerListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allLoadBalancers, err = client.LoadBalancer.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}
		if len(allLoadBalancers) == 0 {
			return fmt.Errorf("no Load Balancer found for selector %q", selector)
		}
		if len(allLoadBalancers) > 1 {
			return fmt.Errorf("more than one Load Balancer found for selector %q", selector)
		}
		setLoadBalancerSchema(d, allLoadBalancers[0])
		return
	}
	return fmt.Errorf("please specify an id, a name or a selector to lookup the Load Balancer")
}
