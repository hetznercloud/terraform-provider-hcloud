package loadbalancer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ServiceResourceType is the type name of the Hetzner Cloud Load Balancer
// resource.
const ServiceResourceType = "hcloud_load_balancer_service"

// ServiceResource creates a Terraform schema for the
// hcloud_load_balancer_service resource.
func ServiceResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLoadBalancerServiceCreate,
		ReadContext:   resourceLoadBalancerServiceRead,
		UpdateContext: resourceLoadBalancerServiceUpdate,
		DeleteContext: resourceLoadBalancerServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"http",
					"https",
					"tcp",
				}, false),
			},
			"listen_port": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"destination_port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"proxyprotocol": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"http": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sticky_sessions": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"cookie_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"cookie_lifetime": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"certificates": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
							Computed: true,
						},
						"redirect_http": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"health_check": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"http",
								"https",
								"tcp",
							}, false),
						},
						"port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"interval": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"retries": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"http": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"path": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"response": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tls": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"status_codes": {
										Type:     schema.TypeList,
										Optional: true,
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
		},
	}
}

func resourceLoadBalancerServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var a *hcloud.Action

	c := m.(*hcloud.Client)

	lbID, err := strconv.Atoi(d.Get("load_balancer_id").(string))
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	lb := hcloud.LoadBalancer{ID: lbID}

	protocol := hcloud.LoadBalancerServiceProtocol(d.Get("protocol").(string))
	opts := hcloud.LoadBalancerAddServiceOpts{
		Protocol: protocol,
	}
	listenPort := d.Get("listen_port").(int)
	// listenPort is a computed attribute. Since we are about to read the resource
	// it may not have been set yet. If this is the case we derive it from the
	// protocol
	if listenPort == 0 {
		switch protocol {
		case hcloud.LoadBalancerServiceProtocolHTTP:
			listenPort = 80
		case hcloud.LoadBalancerServiceProtocolHTTPS:
			listenPort = 443
		}
	}
	opts.ListenPort = hcloud.Int(listenPort)
	if p, ok := d.GetOk("destination_port"); ok {
		opts.DestinationPort = hcloud.Int(p.(int))
	}
	if pp, ok := d.GetOk("proxyprotocol"); ok {
		opts.Proxyprotocol = hcloud.Bool(pp.(bool))
	}
	if tfHTTP, ok := d.GetOk("http"); ok {
		opts.HTTP = parseTFHTTP(tfHTTP.([]interface{}))
	}
	if tfHealthCheck, ok := d.GetOk("health_check"); ok {
		opts.HealthCheck = parseTFHealthCheckAdd(tfHealthCheck.([]interface{}))
	}

	err = control.Retry(control.DefaultRetries, func() error {
		var err error

		a, _, err = c.LoadBalancer.AddService(ctx, &lb, opts)
		if hcloud.IsError(err, hcloud.ErrorCodeServiceError) {
			// Terraform performs CRUD operations for different resources of the
			// same type in parallel. As such it can happen, that a service can't
			// be added, because another service which has not been deleted yet
			// prevents it. We therefore retry the action after a short delay. This
			// should give Terraform enough time to remove the conflicting service
			// (if there is one).
			return err
		}
		return control.AbortRetry(err)
	})
	if resourceLoadBalancerIsNotFound(err, d) {
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	svcID := fmt.Sprintf("%d__%d", lb.ID, listenPort)

	d.SetId(svcID)

	return resourceLoadBalancerServiceRead(ctx, d, m)
}

func resourceLoadBalancerServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	lb, svc, err := lookupLoadBalancerServiceID(ctx, d.Id(), c)
	if err == errInvalidLoadBalancerServiceID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	protocol := hcloud.LoadBalancerServiceProtocol(d.Get("protocol").(string))
	opts := hcloud.LoadBalancerUpdateServiceOpts{
		Protocol: protocol,
	}

	pp := d.Get("proxyprotocol")
	opts.Proxyprotocol = hcloud.Bool(pp.(bool))

	if p, ok := d.GetOk("destination_port"); ok {
		opts.DestinationPort = hcloud.Int(p.(int))
	}

	if tfHTTP, ok := d.GetOk("http"); ok {
		opts.HTTP = parseUpdateTFHTTP(tfHTTP.([]interface{}))
	}

	if tfHealthCheck, ok := d.GetOk("health_check"); ok {
		opts.HealthCheck = parseTFHealthCheckUpdate(tfHealthCheck.([]interface{}))
	}

	action, _, err := c.LoadBalancer.UpdateService(ctx, lb, svc.ListenPort, opts)
	if err != nil {
		if resourceLoadBalancerIsNotFound(err, d) {
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return resourceLoadBalancerServiceRead(ctx, d, m)
}

func resourceLoadBalancerServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	lb, svc, err := lookupLoadBalancerServiceID(ctx, d.Id(), client)
	if err == errInvalidLoadBalancerServiceID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if setLoadBalancerServiceSchema(d, lb, svc); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return nil
}

func resourceLoadBalancerServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	const op = "hcloud/resourceLoadBalancerServiceDelete"

	c := m.(*hcloud.Client)

	lb, svc, err := lookupLoadBalancerServiceID(ctx, d.Id(), c)
	if err == errInvalidLoadBalancerServiceID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	a, _, err := c.LoadBalancer.DeleteService(ctx, lb, svc.ListenPort)
	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		return nil
	}
	if err != nil {
		return diag.Errorf("%s: %v", op, err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, a); err != nil {
		return diag.Errorf("%s: %v", op, err)
	}

	return nil
}

func setLoadBalancerServiceSchema(d *schema.ResourceData, lb *hcloud.LoadBalancer, svc *hcloud.LoadBalancerService) {
	svcID := fmt.Sprintf("%d__%d", lb.ID, svc.ListenPort)

	d.SetId(svcID)
	d.Set("load_balancer_id", strconv.Itoa(lb.ID))
	d.Set("protocol", string(svc.Protocol))
	d.Set("listen_port", svc.ListenPort)
	d.Set("destination_port", svc.DestinationPort)
	d.Set("proxyprotocol", svc.Proxyprotocol)

	if svc.Protocol != hcloud.LoadBalancerServiceProtocolTCP {
		httpMap := make(map[string]interface{})
		if svc.HTTP.StickySessions {
			httpMap["sticky_sessions"] = svc.HTTP.StickySessions
		}
		if svc.HTTP.CookieName != "" {
			httpMap["cookie_name"] = svc.HTTP.CookieName
		}
		if svc.HTTP.CookieLifetime > 0 {
			httpMap["cookie_lifetime"] = int(svc.HTTP.CookieLifetime.Seconds())
		}
		if len(svc.HTTP.Certificates) > 0 {
			certIDs := make([]int, len(svc.HTTP.Certificates))
			for i := 0; i < len(svc.HTTP.Certificates); i++ {
				certIDs[i] = svc.HTTP.Certificates[i].ID
			}
			httpMap["certificates"] = certIDs
		}
		httpMap["redirect_http"] = svc.HTTP.RedirectHTTP
		if len(httpMap) > 0 {
			d.Set("http", []interface{}{httpMap})
		}
	}

	healthCheck := toTFHealthCheck(svc.HealthCheck)
	if len(healthCheck) > 0 {
		d.Set("health_check", []interface{}{healthCheck})
	}
}

var errInvalidLoadBalancerServiceID = errors.New("invalid load balancer service id")

// lookupLoadBalancerServiceID parses the terraform load balancer service record id and return the load balancer and the service
//
// id format: <load balancer id>__<listen-port>
// Examples:
// 123__80
func lookupLoadBalancerServiceID(
	ctx context.Context,
	terraformID string,
	client *hcloud.Client,
) (*hcloud.LoadBalancer, *hcloud.LoadBalancerService, error) {
	if terraformID == "" {
		return nil, nil, errInvalidLoadBalancerServiceID
	}
	parts := strings.SplitN(terraformID, "__", 2)
	if len(parts) != 2 {
		return nil, nil, errInvalidLoadBalancerServiceID
	}

	loadBalancerID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, nil, errInvalidLoadBalancerServiceID
	}

	loadBalancer, _, err := client.LoadBalancer.GetByID(ctx, loadBalancerID)
	if loadBalancer == nil || err != nil {
		return nil, nil, errInvalidLoadBalancerServiceID
	}

	serviceListenPort, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil, errInvalidLoadBalancerServiceID
	}

	for _, svc := range loadBalancer.Services {
		if svc.ListenPort == serviceListenPort {
			svc := svc
			return loadBalancer, &svc, nil
		}
	}
	return nil, nil, errInvalidLoadBalancerServiceID
}

func parseTFHTTP(tfHTTP []interface{}) *hcloud.LoadBalancerAddServiceOptsHTTP {
	if len(tfHTTP) != 1 {
		return nil
	}
	httpMap := tfHTTP[0].(map[string]interface{})
	if len(httpMap) == 0 {
		return nil
	}
	http := &hcloud.LoadBalancerAddServiceOptsHTTP{}
	if stickySessions, ok := httpMap["sticky_sessions"]; ok {
		http.StickySessions = hcloud.Bool(stickySessions.(bool))
	}
	if cookieName, ok := httpMap["cookie_name"]; ok && cookieName != "" {
		http.CookieName = hcloud.String(cookieName.(string))
	}
	if cookieLifetime, ok := httpMap["cookie_lifetime"]; ok && cookieLifetime != 0 {
		http.CookieLifetime = hcloud.Duration(time.Duration(cookieLifetime.(int)) * time.Second)
	}

	if certificates, ok := httpMap["certificates"]; ok {
		http.Certificates = parseTFCertificates(certificates.(*schema.Set))
	}
	if redirectHTTP, ok := httpMap["redirect_http"]; ok {
		http.RedirectHTTP = hcloud.Bool(redirectHTTP.(bool))
	}
	return http
}

func parseUpdateTFHTTP(tfHTTP []interface{}) *hcloud.LoadBalancerUpdateServiceOptsHTTP {
	if len(tfHTTP) != 1 {
		return nil
	}
	httpMap := tfHTTP[0].(map[string]interface{})
	if len(httpMap) == 0 {
		return nil
	}
	http := &hcloud.LoadBalancerUpdateServiceOptsHTTP{}
	if stickySessions, ok := httpMap["sticky_sessions"]; ok {
		http.StickySessions = hcloud.Bool(stickySessions.(bool))
	}
	if cookieName, ok := httpMap["cookie_name"]; ok {
		http.CookieName = hcloud.String(cookieName.(string))
	}
	if cookieLifetime, ok := httpMap["cookie_lifetime"]; ok {
		http.CookieLifetime = hcloud.Duration(time.Duration(cookieLifetime.(int)) * time.Second)
	}

	if certificates, ok := httpMap["certificates"]; ok {
		http.Certificates = parseTFCertificates(certificates.(*schema.Set))
	}
	if redirectHTTP, ok := httpMap["redirect_http"]; ok {
		http.RedirectHTTP = hcloud.Bool(redirectHTTP.(bool))
	}
	return http
}

func parseTFCertificates(tfCerts *schema.Set) []*hcloud.Certificate {
	if tfCerts.Len() == 0 {
		return nil
	}
	certs := make([]*hcloud.Certificate, tfCerts.Len())
	for i, c := range tfCerts.List() {
		certs[i] = &hcloud.Certificate{ID: c.(int)}
	}
	return certs
}

func toTFHealthCheck(healthCheck hcloud.LoadBalancerServiceHealthCheck) map[string]interface{} {
	healthCheckMap := make(map[string]interface{})

	healthCheckMap["protocol"] = healthCheck.Protocol
	healthCheckMap["port"] = healthCheck.Port
	healthCheckMap["interval"] = healthCheck.Interval / time.Second
	healthCheckMap["timeout"] = healthCheck.Timeout / time.Second
	if healthCheck.Retries > 0 {
		healthCheckMap["retries"] = healthCheck.Retries
	}
	if healthCheck.HTTP != nil {
		httpMap := make(map[string]interface{})

		if healthCheck.HTTP.Domain != "" {
			httpMap["domain"] = healthCheck.HTTP.Domain
		}
		if healthCheck.HTTP.Path != "" {
			httpMap["path"] = healthCheck.HTTP.Path
		}
		if healthCheck.HTTP.Response != "" {
			httpMap["response"] = healthCheck.HTTP.Response
		}
		httpMap["tls"] = healthCheck.HTTP.TLS
		httpMap["status_codes"] = healthCheck.HTTP.StatusCodes

		healthCheckMap["http"] = []interface{}{httpMap}
	}

	return healthCheckMap
}

func parseTFHealthCheckAdd(tfHealthCheck []interface{}) *hcloud.LoadBalancerAddServiceOptsHealthCheck {
	var healthCheckOpts hcloud.LoadBalancerAddServiceOptsHealthCheck

	if len(tfHealthCheck) != 1 {
		return nil
	}
	healthCheckMap := tfHealthCheck[0].(map[string]interface{})
	healthCheckOpts.Protocol = hcloud.LoadBalancerServiceProtocol(healthCheckMap["protocol"].(string))
	if port, ok := healthCheckMap["port"]; ok {
		healthCheckOpts.Port = hcloud.Int(port.(int))
	}
	if interval, ok := healthCheckMap["interval"]; ok {
		healthCheckOpts.Interval = hcloud.Duration(time.Duration(interval.(int)) * time.Second)
	}
	if timeout, ok := healthCheckMap["timeout"]; ok {
		healthCheckOpts.Timeout = hcloud.Duration(time.Duration(timeout.(int)) * time.Second)
	}
	if retries, ok := healthCheckMap["retries"]; ok {
		healthCheckOpts.Retries = hcloud.Int(retries.(int))
	}
	if http, ok := healthCheckMap["http"]; ok {
		healthCheckOpts.HTTP = parseTFHealthCheckHTTPAdd(http.([]interface{}))
	}

	return &healthCheckOpts
}

func parseTFHealthCheckUpdate(tfHealthCheck []interface{}) *hcloud.LoadBalancerUpdateServiceOptsHealthCheck {
	var healthCheckOpts hcloud.LoadBalancerUpdateServiceOptsHealthCheck

	if len(tfHealthCheck) != 1 {
		return nil
	}
	healthCheckMap := tfHealthCheck[0].(map[string]interface{})
	healthCheckOpts.Protocol = hcloud.LoadBalancerServiceProtocol(healthCheckMap["protocol"].(string))
	if port, ok := healthCheckMap["port"]; ok {
		healthCheckOpts.Port = hcloud.Int(port.(int))
	}
	if interval, ok := healthCheckMap["interval"]; ok {
		healthCheckOpts.Interval = hcloud.Duration(time.Duration(interval.(int)) * time.Second)
	}
	if timeout, ok := healthCheckMap["timeout"]; ok {
		healthCheckOpts.Timeout = hcloud.Duration(time.Duration(timeout.(int)) * time.Second)
	}
	if retries, ok := healthCheckMap["retries"]; ok {
		healthCheckOpts.Retries = hcloud.Int(retries.(int))
	}
	if http, ok := healthCheckMap["http"]; ok {
		healthCheckOpts.HTTP = parseTFHealthCheckHTTPUpdate(http.([]interface{}))
	}

	return &healthCheckOpts
}

func parseTFHealthCheckHTTPAdd(tfHealthCheckHTTP []interface{}) *hcloud.LoadBalancerAddServiceOptsHealthCheckHTTP {
	if len(tfHealthCheckHTTP) != 1 {
		return nil
	}
	httpMap := tfHealthCheckHTTP[0].(map[string]interface{})
	httpHealthCheck := &hcloud.LoadBalancerAddServiceOptsHealthCheckHTTP{}

	if domain, ok := httpMap["domain"]; ok {
		httpHealthCheck.Domain = hcloud.String(domain.(string))
	}
	if path, ok := httpMap["path"]; ok {
		httpHealthCheck.Path = hcloud.String(path.(string))
	}
	if response, ok := httpMap["response"]; ok {
		httpHealthCheck.Response = hcloud.String(response.(string))
	}
	if tls, ok := httpMap["tls"]; ok {
		httpHealthCheck.TLS = hcloud.Bool(tls.(bool))
	}
	if scs, ok := httpMap["status_codes"]; ok {
		var statusCodes []string

		for _, sc := range scs.([]interface{}) {
			statusCodes = append(statusCodes, sc.(string))
		}
		httpHealthCheck.StatusCodes = statusCodes
	}
	return httpHealthCheck
}

func parseTFHealthCheckHTTPUpdate(tfHealthCheckHTTP []interface{}) *hcloud.LoadBalancerUpdateServiceOptsHealthCheckHTTP {
	if len(tfHealthCheckHTTP) != 1 {
		return nil
	}
	httpMap := tfHealthCheckHTTP[0].(map[string]interface{})
	httpHealthCheck := &hcloud.LoadBalancerUpdateServiceOptsHealthCheckHTTP{}

	if domain, ok := httpMap["domain"]; ok {
		httpHealthCheck.Domain = hcloud.String(domain.(string))
	}
	if path, ok := httpMap["path"]; ok {
		httpHealthCheck.Path = hcloud.String(path.(string))
	}
	if response, ok := httpMap["response"]; ok {
		httpHealthCheck.Response = hcloud.String(response.(string))
	}
	if tls, ok := httpMap["tls"]; ok {
		httpHealthCheck.TLS = hcloud.Bool(tls.(bool))
	}
	if scs, ok := httpMap["status_codes"]; ok {
		var statusCodes []string

		for _, sc := range scs.([]interface{}) {
			statusCodes = append(statusCodes, sc.(string))
		}
		httpHealthCheck.StatusCodes = statusCodes
	}
	return httpHealthCheck
}
