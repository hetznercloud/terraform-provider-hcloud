package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudSSHKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudSSHKeysRead,
		Schema: map[string]*schema.Schema{
			"with_selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fingerprint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"labels": {
							Type:     schema.TypeMap,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
func dataSourceHcloudSSHKeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	labelSelector := d.Get("with_selector")
	labelSelectorStr, _ := labelSelector.(string)

	opts := hcloud.SSHKeyListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: labelSelectorStr,
		},
	}
	keys, err := client.SSHKey.AllWithOpts(context.Background(), opts)
	if err != nil {
		return diag.FromErr(err)
	}
	var keyMaps []map[string]interface{}
	id := ""
	for _, key := range keys {
		if id != "" {
			id += "-"
		}
		id += fmt.Sprintf("%d", key.ID)
		keyMaps = append(keyMaps, map[string]interface{}{
			"id":          key.ID,
			"name":        key.Name,
			"fingerprint": key.Fingerprint,
			"public_key":  key.PublicKey,
			"labels":      key.Labels,
		})
	}

	d.SetId(id)
	d.Set("ssh_keys", keyMaps)
	return nil
}
