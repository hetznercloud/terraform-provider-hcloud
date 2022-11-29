package firewall

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// AttachmentResourceType is the type of the hcloud_firewall_attachment resource.
const AttachmentResourceType = "hcloud_firewall_attachment"

// AttachmentResource defines the schema for the hcloud_firewall_attachment
// resource.
func AttachmentResource() *schema.Resource {
	return &schema.Resource{
		ReadContext:   readAttachment,
		CreateContext: createAttachment,
		UpdateContext: updateAttachment,
		DeleteContext: deleteAttachment,

		Schema: map[string]*schema.Schema{
			"firewall_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"server_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"label_selectors": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func readAttachment(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var att attachment

	if err := att.FromResourceData(d); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*hcloud.Client)
	fw, _, err := client.Firewall.GetByID(ctx, att.FirewallID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if fw == nil {
		log.Printf("[WARN] firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := att.FromFirewall(fw); err != nil {
		return diag.FromErr(err)
	}
	att.ToResourceData(d)

	return nil
}

func createAttachment(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	var att attachment

	if err := att.FromResourceData(d); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*hcloud.Client)
	action, _, err := client.Firewall.ApplyResources(ctx, &hcloud.Firewall{ID: att.FirewallID}, att.AllResources())
	if hcloud.IsError(err, hcloud.ErrorCodeFirewallAlreadyApplied) {
		return readAttachment(ctx, d, m)
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForActions(ctx, &client.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return readAttachment(ctx, d, m)
}

func updateAttachment(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		tf, hc  attachment
		actions []*hcloud.Action
	)

	if err := tf.FromResourceData(d); err != nil {
		return diag.FromErr(err)
	}

	client := m.(*hcloud.Client)
	fw, _, err := client.Firewall.GetByID(ctx, tf.FirewallID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if fw == nil {
		log.Printf("[WARN] firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err := hc.FromFirewall(fw); err != nil {
		return diag.FromErr(err)
	}

	less, more := tf.DiffResources(hc)
	as, _, err := client.Firewall.RemoveResources(ctx, fw, less)
	if err != nil && !hcloud.IsError(err, hcloud.ErrorCodeFirewallAlreadyRemoved) {
		return hcclient.ErrorToDiag(err)
	}
	actions = append(actions, as...)

	as, _, err = client.Firewall.ApplyResources(ctx, fw, more)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	actions = append(actions, as...)

	if err := hcclient.WaitForActions(ctx, &client.Action, actions); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return readAttachment(ctx, d, m)
}

func deleteAttachment(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	var att attachment

	defer func() {
		if diags != nil {
			return
		}
		d.SetId("")
	}()

	if err := att.FromResourceData(d); err != nil {
		return diag.FromErr(err)
	}
	client := m.(*hcloud.Client)
	action, _, err := client.Firewall.RemoveResources(ctx, &hcloud.Firewall{ID: att.FirewallID}, att.AllResources())
	if hcloud.IsError(err, hcloud.ErrorCodeFirewallAlreadyRemoved) {
		return nil
	}
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForActions(ctx, &client.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return nil
}

type attachment struct {
	FirewallID     int
	ServerIDs      []int
	LabelSelectors []string
}

// FromResourceData copies the contents of d into a
func (a *attachment) FromResourceData(d *schema.ResourceData) error {
	// The terraform schema definition above ensures this is always set and
	// of the correct type. Thus there is no need to check such things.
	a.FirewallID = d.Get("firewall_id").(int)

	srvIDs, ok := d.GetOk("server_ids")
	if ok {
		for _, v := range srvIDs.(*schema.Set).List() {
			a.ServerIDs = append(a.ServerIDs, v.(int))
		}
		sort.Slice(a.ServerIDs, func(i, j int) bool {
			return a.ServerIDs[i] < a.ServerIDs[j]
		})
	}

	lSels, ok := d.GetOk("label_selectors")
	if ok {
		for _, v := range lSels.(*schema.Set).List() {
			a.LabelSelectors = append(a.LabelSelectors, v.(string))
		}
		sort.Slice(a.LabelSelectors, func(i, j int) bool {
			return a.LabelSelectors[i] < a.LabelSelectors[j]
		})
	}

	if len(a.ServerIDs) == 0 && len(a.LabelSelectors) == 0 {
		return fmt.Errorf("no resources referenced")
	}
	return nil
}

// ToResourceData copies the contents of a into d.
//
// Any previously existing values in d are overwritten or removed.
func (a *attachment) ToResourceData(d *schema.ResourceData) {
	var srvIDs, lSels *schema.Set

	if len(a.ServerIDs) > 0 {
		vals := make([]interface{}, len(a.ServerIDs))
		for i, id := range a.ServerIDs {
			vals[i] = id
		}
		f := d.Get("server_ids").(*schema.Set).F // Returns a default value if server_ids is not present in HCL.
		srvIDs = schema.NewSet(f, vals)
	}
	d.Set("server_ids", srvIDs)

	if len(a.LabelSelectors) > 0 {
		vals := make([]interface{}, len(a.LabelSelectors))
		for i, ls := range a.LabelSelectors {
			vals[i] = ls
		}
		f := d.Get("label_selectors").(*schema.Set).F
		lSels = schema.NewSet(f, vals)
	}
	d.Set("label_selectors", lSels)

	d.Set("firewall_id", a.FirewallID)
	d.SetId(strconv.Itoa(a.FirewallID))
}

// FromFirewall reads the attachment data from fw into a.
func (a *attachment) FromFirewall(fw *hcloud.Firewall) error {
	// We do not need to read the fw.ID. This value is always set in HCL.
	// Additionally the intended use-case makes sure that we never get data
	// for the wrong firewall. Therefore comparing fw.ID to a.FirewallID
	// is not necessary.

	for _, fwr := range fw.AppliedTo {
		switch fwr.Type {
		case hcloud.FirewallResourceTypeServer:
			a.ServerIDs = append(a.ServerIDs, fwr.Server.ID)
		case hcloud.FirewallResourceTypeLabelSelector:
			a.LabelSelectors = append(a.LabelSelectors, fwr.LabelSelector.Selector)
		default:
			return fmt.Errorf("invalid firewall resource type: %v", fwr.Type)
		}
	}

	return nil
}

// AllResources returns all Hetzner Cloud Firewall that are attached to
// the Firewall of this attachment.
func (a *attachment) AllResources() []hcloud.FirewallResource {
	n := len(a.ServerIDs) + len(a.LabelSelectors)
	if n == 0 {
		return nil
	}

	ress := make([]hcloud.FirewallResource, 0, n)
	for _, id := range a.ServerIDs {
		ress = append(ress, serverResource(id))
	}
	for _, ls := range a.LabelSelectors {
		ress = append(ress, labelSelectorResource(ls))
	}

	return ress
}

// DiffResources compares the Firewall resources of a to the resources of o.
//
// The first return value contains all resources that are present in o but
// missing in a. The second return value is a slice containing all resources
// present in a but missing in o.
func (a *attachment) DiffResources(o attachment) ([]hcloud.FirewallResource, []hcloud.FirewallResource) {
	var more, less []hcloud.FirewallResource // nolint: prealloc

	aSrvs := make(map[int]bool, len(a.ServerIDs))
	for _, id := range a.ServerIDs {
		aSrvs[id] = true
	}
	for _, id := range o.ServerIDs {
		if aSrvs[id] {
			continue
		}
		less = append(less, serverResource(id))
	}

	aLSels := make(map[string]bool, len(a.LabelSelectors))
	for _, ls := range a.LabelSelectors {
		aLSels[ls] = true
	}
	for _, ls := range o.LabelSelectors {
		if aLSels[ls] {
			continue
		}
		less = append(less, labelSelectorResource(ls))
	}

	oSrvs := make(map[int]bool, len(o.ServerIDs))
	for _, id := range o.ServerIDs {
		oSrvs[id] = true
	}
	for _, id := range a.ServerIDs {
		if oSrvs[id] {
			continue
		}
		more = append(more, serverResource(id))
	}

	oLSels := make(map[string]bool, len(o.LabelSelectors))
	for _, ls := range o.LabelSelectors {
		oLSels[ls] = true
	}
	for _, ls := range a.LabelSelectors {
		if oLSels[ls] {
			continue
		}
		more = append(more, labelSelectorResource(ls))
	}

	return less, more
}

func serverResource(id int) hcloud.FirewallResource {
	return hcloud.FirewallResource{
		Type:   hcloud.FirewallResourceTypeServer,
		Server: &hcloud.FirewallResourceServer{ID: id},
	}
}

func labelSelectorResource(ls string) hcloud.FirewallResource {
	return hcloud.FirewallResource{
		Type:          hcloud.FirewallResourceTypeLabelSelector,
		LabelSelector: &hcloud.FirewallResourceLabelSelector{Selector: ls},
	}
}
