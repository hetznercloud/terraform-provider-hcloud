package firewall

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud Firewall resource.
const ResourceType = "hcloud_firewall"

func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFirewallCreate,
		ReadContext:   resourceFirewallRead,
		UpdateContext: resourceFirewallUpdate,
		DeleteContext: resourceFirewallDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"direction": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								direction := i.(string)
								switch hcloud.FirewallRuleDirection(direction) {
								case hcloud.FirewallRuleDirectionIn:
								case hcloud.FirewallRuleDirectionOut:
									return nil
								default:
									return diag.Errorf("%s is not a valid direction", direction)
								}
								return nil
							},
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								protocol := i.(string)
								switch hcloud.FirewallRuleProtocol(protocol) {
								case hcloud.FirewallRuleProtocolICMP:
								case hcloud.FirewallRuleProtocolTCP:
								case hcloud.FirewallRuleProtocolUDP:
								case hcloud.FirewallRuleProtocolESP:
								case hcloud.FirewallRuleProtocolGRE:
									return nil
								default:
									return diag.Errorf("%s is not a valid protocol", protocol)
								}
								return nil
							},
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_ips": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validateIPDiag,
							},
							Optional: true,
						},
						"destination_ips": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validateIPDiag,
							},
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceFirewallCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	opts := hcloud.FirewallCreateOpts{
		Name: d.Get("name").(string),
	}
	if rules, ok := d.GetOk("rule"); ok {
		for _, tfRawRule := range rules.(*schema.Set).List() {
			rule := toHcloudRule(tfRawRule)
			opts.Rules = append(opts.Rules, rule)
		}
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	res, _, err := client.Firewall.Create(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	for _, nextAction := range res.Actions {
		if err := hcclient.WaitForAction(ctx, &client.Action, nextAction); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	d.SetId(strconv.Itoa(res.Firewall.ID))

	return resourceFirewallRead(ctx, d, m)
}

func toHcloudRule(tfRawRule interface{}) hcloud.FirewallRule {
	tfRule := tfRawRule.(map[string]interface{})
	rule := hcloud.FirewallRule{
		Direction: hcloud.FirewallRuleDirection(tfRule["direction"].(string)),
		Protocol:  hcloud.FirewallRuleProtocol(tfRule["protocol"].(string)),
	}
	rawPort := tfRule["port"].(string)
	if rawPort != "" {
		rule.Port = hcloud.String(rawPort)
	}
	for _, sourceIP := range tfRule["source_ips"].(*schema.Set).List() {
		// We ignore the error here, because it was already validated before
		_, source, _ := net.ParseCIDR(sourceIP.(string))
		rule.SourceIPs = append(rule.SourceIPs, *source)
	}
	for _, destinationIP := range tfRule["destination_ips"].(*schema.Set).List() {
		// We ignore the error here, because it was already validated before
		_, destination, _ := net.ParseCIDR(destinationIP.(string))
		rule.DestinationIPs = append(rule.DestinationIPs, *destination)
	}
	return rule
}

func toTFRule(hcloudRule hcloud.FirewallRule) map[string]interface{} {
	tfRule := make(map[string]interface{})
	tfRule["direction"] = string(hcloudRule.Direction)
	tfRule["protocol"] = string(hcloudRule.Protocol)

	if hcloudRule.Port != nil {
		tfRule["port"] = hcloudRule.Port
	}
	sourceIPs := make([]string, len(hcloudRule.SourceIPs))
	for i, sourceIP := range hcloudRule.SourceIPs {
		sourceIPs[i] = sourceIP.String()
	}
	tfRule["source_ips"] = sourceIPs
	destinationIPs := make([]string, len(hcloudRule.DestinationIPs))
	for i, destinationIP := range hcloudRule.DestinationIPs {
		destinationIPs[i] = destinationIP.String()
	}
	tfRule["destination_ips"] = destinationIPs
	return tfRule
}

func resourceFirewallRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid firewall id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	firewall, _, err := client.Firewall.GetByID(ctx, id)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if firewall == nil {
		log.Printf("[WARN] firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setFirewallSchema(d, firewall)
	return nil
}

func resourceFirewallUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid firewall id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	firewall, _, err := client.Firewall.GetByID(ctx, id)
	if err != nil {
		if resourceFirewallIsNotFound(err, d) {
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	d.Partial(true)

	if d.HasChange("name") {
		description := d.Get("name").(string)
		_, _, err := client.Firewall.Update(ctx, firewall, hcloud.FirewallUpdateOpts{
			Name: description,
		})
		if err != nil {
			if resourceFirewallIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("rule") {
		if tfRules, ok := d.GetOk("rule"); ok {
			var rules []hcloud.FirewallRule
			for _, tfRawRule := range tfRules.(*schema.Set).List() {
				rule := toHcloudRule(tfRawRule)
				rules = append(rules, rule)
			}
			actions, _, err := client.Firewall.SetRules(ctx, firewall, hcloud.FirewallSetRulesOpts{Rules: rules})
			if err != nil {
				if resourceFirewallIsNotFound(err, d) {
					return nil
				}
				return hcclient.ErrorToDiag(err)
			}
			if err := waitForFirewallActions(ctx, client, actions, firewall); err != nil {
				return hcclient.ErrorToDiag(err)
			}
		}
	}

	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.Firewall.Update(ctx, firewall, hcloud.FirewallUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceFirewallIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}
	d.Partial(false)

	return resourceFirewallRead(ctx, d, m)
}

func resourceFirewallDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	firewallID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid firewall id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	firewall, _, err := client.Firewall.GetByID(ctx, firewallID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	err = control.Retry(control.DefaultRetries, func() error {
		var hcerr hcloud.Error
		_, err := client.Firewall.Delete(ctx, firewall)
		if errors.As(err, &hcerr) {
			switch hcerr.Code {
			case hcloud.ErrorCodeNotFound:
				// firewall has already been deleted
				return nil
			case hcloud.ErrorCodeConflict, hcloud.ErrorCodeResourceInUse:
				return err
			default:
				return control.AbortRetry(err)
			}
		}
		return nil
	})
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func resourceFirewallIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setFirewallSchema(d *schema.ResourceData, v *hcloud.Firewall) {
	d.SetId(strconv.Itoa(v.ID))
	d.Set("name", v.Name)

	rules := make([]map[string]interface{}, len(v.Rules))
	for i, rule := range v.Rules {
		rules[i] = toTFRule(rule)
	}
	d.Set("rule", rules)
	d.Set("labels", v.Labels)
}

func waitForFirewallActions(ctx context.Context, client *hcloud.Client, actions []*hcloud.Action, firewall *hcloud.Firewall) error {
	log.Printf("[INFO] firewall (%d) waiting for %v actions to complete...", firewall.ID, len(actions))
	_, errCh := client.Action.WatchOverallProgress(ctx, actions)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] firewall (%d) %v actions succeeded", firewall.ID, len(actions))
	return nil
}
