package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudSSHKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudSSHKeyRead,
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
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}
func dataSourceHcloudSSHKeyRead(d *schema.ResourceData, m interface{}) (err error) {
	client := m.(*hcloud.Client)
	ctx := context.Background()
	var s *hcloud.SSHKey
	if id, ok := d.GetOk("id"); ok {
		s, _, err = client.SSHKey.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if s == nil {
			return fmt.Errorf("no sshkey found with id %d", id)
		}
		setSSHKeySchema(d, s)
		return
	}
	if name, ok := d.GetOk("name"); ok {
		s, _, err = client.SSHKey.GetByName(ctx, name.(string))
		if err != nil {
			return err
		}
		if s == nil {
			return fmt.Errorf("no sshkey found with name %v", name)
		}
		setSSHKeySchema(d, s)
		return
	}
	if fingerprint, ok := d.GetOk("fingerprint"); ok {
		s, _, err = client.SSHKey.GetByFingerprint(ctx, fingerprint.(string))
		if err != nil {
			return err
		}
		if s == nil {
			return fmt.Errorf("no sshkey found with name %v", fingerprint)
		}
		setSSHKeySchema(d, s)
		return
	}
	if selector, ok := d.GetOk("selector"); ok {
		var allKeys []*hcloud.SSHKey
		opts := hcloud.SSHKeyListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector.(string),
			},
		}
		allKeys, err = client.SSHKey.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}
		if len(allKeys) == 0 {
			return fmt.Errorf("no sshkey found for selector %q", selector)
		}
		if len(allKeys) > 1 {
			return fmt.Errorf("more than one sshkey found for selector %q", selector)
		}
		setSSHKeySchema(d, allKeys[0])
		return
	}
	return fmt.Errorf("please specify a id, a name, a fingerprint or a selector to lookup the sshkey")
}
