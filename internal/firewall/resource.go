package firewall

import (
	"context"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
	"log"
	"net"
	"strconv"
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
						"direction": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								direction := i.(string)
								switch hcloud.FirewallRuleDirection(direction) {
								case hcloud.FirewallRuleDirectionIn:
									return nil
								default:
									return diag.Errorf("%s is not a valid direction", direction)
								}
							},
						},
						"protocol": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								protocol := i.(string)
								switch hcloud.FirewallRuleProtocol(protocol) {
								case hcloud.FirewallRuleProtocolICMP:
								case hcloud.FirewallRuleProtocolTCP:
								case hcloud.FirewallRuleProtocolUDP:
									return nil
								default:
									return diag.Errorf("%s is not a valid protocol", protocol)
								}
								return nil
							},
						},
						"port": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_ips": &schema.Schema{
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
									sourceIP := i.(string)
									_, _, err := net.ParseCIDR(sourceIP)
									if err != nil {
										return diag.FromErr(err)
									}
									return nil
								},
							},
							Required: true,
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
		return diag.FromErr(err)
	}

	for _, nextAction := range res.Actions {
		if err := hcclient.WaitForAction(ctx, &client.Action, nextAction); err != nil {
			return diag.FromErr(err)
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
	for _, sourceIP := range tfRule["source_ips"].([]interface{}) {
		// We ignore the error here, because it was already validated before
		_, source, _ := net.ParseCIDR(sourceIP.(string))
		rule.SourceIPs = append(rule.SourceIPs, *source)
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
	var sourceIPs []string
	for _, sourceIP := range hcloudRule.SourceIPs {
		sourceIPs = append(sourceIPs, sourceIP.String())
	}
	tfRule["source_ips"] = sourceIPs
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
		return diag.FromErr(err)
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
		return diag.FromErr(err)
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
			return diag.FromErr(err)
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
				return diag.FromErr(err)
			}
			if err := waitForFirewallActions(ctx, client, actions, firewall); err != nil {
				return diag.FromErr(err)
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
			return diag.FromErr(err)
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
		return diag.FromErr(err)
	}

	if _, err := client.Firewall.Delete(ctx, firewall); err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// firewall has already been deleted
			return nil
		}
		return diag.FromErr(err)
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

	var rules []map[string]interface{}
	for _, rule := range v.Rules {
		rules = append(rules, toTFRule(rule))
	}
	d.Set("rule", rules)
	d.Set("labels", v.Labels)
}

func waitForFirewallAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action, firewall *hcloud.Firewall) error {
	log.Printf("[INFO] firewall (%d) waiting for %q action to complete...", firewall.ID, action.Command)
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] firewall (%d) %q action succeeded", firewall.ID, action.Command)
	return nil
}

func waitForFirewallActions(ctx context.Context, client *hcloud.Client, actions []*hcloud.Action, firewall *hcloud.Firewall) error {
	log.Printf("[INFO] firewall (%d) waiting for %q action to complete...", firewall.ID, len(actions))
	_, errCh := client.Action.WatchOverallProgress(ctx, actions)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] firewall (%d) %q actions succeeded", firewall.ID, len(actions))
	return nil
}
