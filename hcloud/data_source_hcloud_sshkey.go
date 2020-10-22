package hcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudSSHKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudSSHKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"selector": {
				Type:          schema.TypeString,
				Optional:      true,
				Deprecated:    "Please use the with_selector property instead.",
				ConflictsWith: []string{"with_selector"},
			},
			"with_selector": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"selector"},
			},
		},
	}
}
func dataSourceHcloudSSHKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		s, _, err := client.SSHKey.GetByID(ctx, id.(int))
		if err != nil {
			return diag.FromErr(err)
		}
		if s == nil {
			return diag.Errorf("no sshkey found with id %d", id)
		}
		setSSHKeySchema(d, s)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		s, _, err := client.SSHKey.GetByName(ctx, name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if s == nil {
			return diag.Errorf("no sshkey found with name %v", name)
		}
		setSSHKeySchema(d, s)
		return nil
	}
	if fingerprint, ok := d.GetOk("fingerprint"); ok {
		s, _, err := client.SSHKey.GetByFingerprint(ctx, fingerprint.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if s == nil {
			return diag.Errorf("no sshkey found with fingerprint %v", fingerprint)
		}
		setSSHKeySchema(d, s)
		return nil
	}

	var selector string
	if v := d.Get("with_selector").(string); v != "" {
		selector = v
	} else if v := d.Get("selector").(string); v != "" {
		selector = v
	}
	if selector != "" {
		var allKeys []*hcloud.SSHKey
		opts := hcloud.SSHKeyListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector,
			},
		}
		allKeys, err := client.SSHKey.AllWithOpts(ctx, opts)
		if err != nil {
			return diag.FromErr(err)
		}
		if len(allKeys) == 0 {
			return diag.Errorf("no sshkey found for selector %q", selector)
		}
		if len(allKeys) > 1 {
			return diag.Errorf("more than one sshkey found for selector %q", selector)
		}
		setSSHKeySchema(d, allKeys[0])
		return nil
	}
	return diag.Errorf("please specify a id, a name, a fingerprint or a selector to lookup the sshkey")
}
