package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	// DataSourceServiceType is the type name of the Hetzner Cloud Load Balancer Service resource.
	DataSourceServiceType = "hcloud_load_balancer_service"

	// DataSourceServiceListType is the type name to receive a list of Hetzner Cloud Load Balancer Service resources.
	DataSourceServiceListType = "hcloud_load_balancer_services"
)

func getCommonServiceDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"load_balancer_id": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"protocol": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"listen_port": {
			Type:     schema.TypeInt,
			Computed: true,
			Optional: true,
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
					"timeout_idle": {
						Type:     schema.TypeInt,
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
										Type: schema.TypeString,
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

// DataSourceService creates a new Terraform schema for the hcloud_load_balancer_service
// resource.
func DataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLoadBalancerServiceRead,
		Schema:      getCommonServiceDataSchema(),
	}
}

func DataSourceServiceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudLoadBalancerServiceListRead,
		Schema: map[string]*schema.Schema{
			"service": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonServiceDataSchema(),
				},
			},
			"load_balancer_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func dataSourceHcloudLoadBalancerServiceRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*hcloud.Client)

	var id string
	if _id, ok := d.GetOk("id"); ok {
		id = _id.(string)
	}

	var (
		lbID       int64
		listenPort int
	)
	if _lbID, ok := d.GetOk("load_balancer_id"); ok {
		lbID = util.CastInt64(_lbID)
	}
	if _listenPort, ok := d.GetOk("listen_port"); ok {
		listenPort = _listenPort.(int)
	}
	if lbID > 0 && listenPort > 0 {
		id = fmt.Sprintf("%d__%d", lbID, listenPort)
	}

	if id != "" {
		lb, svc, err := lookupLoadBalancerServiceID(ctx, id, client)
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		setLoadBalancerServiceSchema(d, lb, svc, false)
		return nil
	}

	return diag.Errorf("please specify an id or a load_balancer_id and listen_port to lookup the Load Balancer")
}

func dataSourceHcloudLoadBalancerServiceListRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := m.(*hcloud.Client)

	var lbID int64
	if _lbID, ok := d.GetOk("load_balancer_id"); ok {
		lbID = util.CastInt64(_lbID)
	}

	lb, _, err := client.LoadBalancer.GetByID(ctx, lbID)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if lb == nil {
		return diag.Errorf("load balancer %d: not found", lbID)
	}

	ids := make([]int, len(lb.Services))
	services := make([]any, len(lb.Services))
	for i, svc := range lb.Services {
		ids[i] = svc.ListenPort
		services[i] = getLoadBalancerServiceAttributes(lb, &svc, false)
	}
	d.SetId(datasourceutil.ListID(ids))
	d.Set("service", services)
	return nil
}
