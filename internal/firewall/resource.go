package firewall

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

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
				Computed: true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					if ok, err := hcloud.ValidateResourceLabels(i.(map[string]interface{})); !ok {
						return diag.Errorf(err.Error())
					}
					return nil
				},
			},
			"apply_to": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
					// Diff is only valid if "network" resource is set in
					// terraform configuration.
					_, ok := d.GetOk("apply_to")
					return !ok // Negate because we do **not** want to suppress the diff.
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label_selector": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"server": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
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
								StateFunc: func(i interface{}) string {
									return strings.ToLower(i.(string))
								},
							},
							Optional: true,
						},
						"destination_ips": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validateIPDiag,
								StateFunc: func(i interface{}) string {
									return strings.ToLower(i.(string))
								},
							},
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
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
			if rule, ok := toHcloudRule(tfRawRule); ok {
				opts.Rules = append(opts.Rules, rule)
			}
		}
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}
	if applyTo, ok := d.GetOk("apply_to"); ok {
		for _, tfApplyToRaw := range applyTo.(*schema.Set).List() {
			tfApplyTo := tfApplyToRaw.(map[string]interface{})

			r, _ := toHcloudFirewallResource(tfApplyTo)
			opts.ApplyTo = append(opts.ApplyTo, r)
		}
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

func toHcloudRule(tfRawRule interface{}) (hcloud.FirewallRule, bool) {
	tfRule := tfRawRule.(map[string]interface{})
	direction := tfRule["direction"].(string)
	protocol := tfRule["protocol"].(string)

	if direction == "" || protocol == "" {
		// We need to use state funcs in the schema above in order to normalize
		// the values for source and destination IPs. However, this triggers
		// Terraform SDK bug #160
		// (https://github.com/hashicorp/terraform-plugin-sdk/issues/160):
		// we get a defunct entry in the rules set. Since we always require
		// protocol and direction to be set, we can just check if they are
		// empty. In this case we can ignore the entry when talking to our API
		// (https://github.com/hashicorp/terraform-plugin-sdk/issues/160#issuecomment-522935697).
		return hcloud.FirewallRule{}, false
	}

	rule := hcloud.FirewallRule{
		Direction: hcloud.FirewallRuleDirection(direction),
		Protocol:  hcloud.FirewallRuleProtocol(protocol),
	}
	rawPort := tfRule["port"].(string)
	if rawPort != "" {
		rule.Port = hcloud.Ptr(rawPort)
	}
	rawDescription := tfRule["description"].(string)
	if rawDescription != "" {
		rule.Description = hcloud.Ptr(rawDescription)
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
	return rule, true
}

func toTFRule(hcloudRule hcloud.FirewallRule) map[string]interface{} {
	tfRule := make(map[string]interface{})
	tfRule["direction"] = string(hcloudRule.Direction)
	tfRule["protocol"] = string(hcloudRule.Protocol)

	if hcloudRule.Port != nil {
		tfRule["port"] = hcloudRule.Port
	}
	if hcloudRule.Description != nil {
		tfRule["description"] = hcloudRule.Description
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
				if rule, ok := toHcloudRule(tfRawRule); ok {
					rules = append(rules, rule)
				}
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

	if d.HasChange("apply_to") {
		err := syncApplyTo(ctx, d, client, firewall)
		if err != nil {
			return err
		}
	}
	d.Partial(false)

	return resourceFirewallRead(ctx, d, m)
}

func syncApplyTo(ctx context.Context, d *schema.ResourceData, client *hcloud.Client, firewall *hcloud.Firewall) diag.Diagnostics {
	o, n := d.GetChange("apply_to")

	diffToRemove := o.(*schema.Set).Difference(n.(*schema.Set))
	diffToAdd := n.(*schema.Set).Difference(o.(*schema.Set))

	removeResources := []hcloud.FirewallResource{}
	addResources := []hcloud.FirewallResource{}
	// We first prepare all changes to then simply apply them
	for _, d := range diffToRemove.List() {
		field := d.(map[string]interface{})
		r, _ := toHcloudFirewallResource(field) // we ignore the error here, as it can not happen because of the validation before
		removeResources = append(removeResources, r)
	}

	for _, d := range diffToAdd.List() {
		field := d.(map[string]interface{})

		r, _ := toHcloudFirewallResource(field) // we ignore the error here, as it can not happen because of the validation before
		addResources = append(addResources, r)
	}

	if len(removeResources) > 0 {
		actions, _, err := client.Firewall.RemoveResources(ctx, firewall, removeResources)
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

	if len(addResources) > 0 {
		actions, _, err := client.Firewall.ApplyResources(ctx, firewall, addResources)
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
	return nil
}

func toHcloudFirewallResource(field map[string]interface{}) (hcloud.FirewallResource, error) {
	var op = "toHcloudFirewallResource"
	if labelSelector, ok := field["label_selector"].(string); ok && labelSelector != "" {
		return hcloud.FirewallResource{Type: hcloud.FirewallResourceTypeLabelSelector, LabelSelector: &hcloud.FirewallResourceLabelSelector{Selector: labelSelector}}, nil
	} else if server, ok := field["server"].(int); ok && server != 0 {
		return hcloud.FirewallResource{Type: hcloud.FirewallResourceTypeServer, Server: &hcloud.FirewallResourceServer{ID: server}}, nil
	}
	return hcloud.FirewallResource{}, fmt.Errorf("%s: unknown apply to resource", op)
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
	// Detach all Resources of the firewall before trying to delete it.
	if len(firewall.AppliedTo) > 0 {
		if err := removeFromResources(ctx, client, d, firewall); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	// Removing resources from the firewall can sometimes take longer. We
	// thus retry two times the number of DefaultRetries.
	err = control.Retry(2*control.DefaultRetries, func() error {
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

func removeFromResources(ctx context.Context, client *hcloud.Client, d *schema.ResourceData, fw *hcloud.Firewall) error {
	actions, _, err := client.Firewall.RemoveResources(ctx, fw, fw.AppliedTo)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeFirewallResourceNotFound) || resourceFirewallIsNotFound(err, d) {
			return nil
		}
		return err
	}

	return waitForFirewallActions(ctx, client, actions, fw)
}

func resourceFirewallIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setFirewallSchema(d *schema.ResourceData, f *hcloud.Firewall) {
	for key, val := range getFirewallAttributes(f) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getFirewallAttributes(f *hcloud.Firewall) map[string]interface{} {
	rules := make([]map[string]interface{}, len(f.Rules))
	for i, rule := range f.Rules {
		rules[i] = toTFRule(rule)
	}

	var applyTo []map[string]interface{}

	for _, a := range f.AppliedTo {
		if a.Type == hcloud.FirewallResourceTypeLabelSelector {
			applyTo = append(applyTo, map[string]interface{}{"label_selector": a.LabelSelector.Selector})
		} else if a.Type == hcloud.FirewallResourceTypeServer {
			applyTo = append(applyTo, map[string]interface{}{"server": a.Server.ID})
		}
	}

	return map[string]interface{}{
		"id":       f.ID,
		"name":     f.Name,
		"rule":     rules,
		"labels":   f.Labels,
		"apply_to": applyTo,
	}
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
