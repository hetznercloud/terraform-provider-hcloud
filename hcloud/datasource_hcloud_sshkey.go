package hcloud

import (
	"context"
	"fmt"
	"strconv"

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
	return fmt.Errorf("please specify a id or name to lookup the sshkey")
}

func setSSHKeySchema(d *schema.ResourceData, s *hcloud.SSHKey) {
	d.SetId(strconv.Itoa(s.ID))
	d.Set("name", s.Name)
	d.Set("fingerprint", s.Fingerprint)
	d.Set("public_key", s.PublicKey)
}
