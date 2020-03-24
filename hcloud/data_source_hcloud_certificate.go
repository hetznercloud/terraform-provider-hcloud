package hcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHcloudCertificateRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"with_selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
			"domain_names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_valid_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_valid_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHcloudCertificateRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	if id, ok := d.GetOk("id"); ok {
		cert, _, err := client.Certificate.GetByID(ctx, id.(int))
		if err != nil {
			return err
		}
		if cert == nil {
			return fmt.Errorf("certificate not found: id: %d", id)
		}
		setCertificateSchema(d, cert)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		cert, _, err := client.Certificate.Get(ctx, name.(string))
		if err != nil {
			return err
		}
		if cert == nil {
			return fmt.Errorf("certificate not found: name: %s", name)
		}
		setCertificateSchema(d, cert)
		return nil
	}
	if selector, ok := d.GetOk("with_selector"); ok && selector != "" {
		var allCertificates []*hcloud.Certificate
		opts := hcloud.CertificateListOpts{
			ListOpts: hcloud.ListOpts{
				LabelSelector: selector.(string),
			},
		}
		allCertificates, err := client.Certificate.AllWithOpts(ctx, opts)
		if err != nil {
			return err
		}
		if len(allCertificates) == 0 {
			return fmt.Errorf("no Certificate found for selector %q", selector)
		}
		if len(allCertificates) > 1 {
			return fmt.Errorf("more than one Certificate found for selector %q", selector)
		}
		setCertificateSchema(d, allCertificates[0])
		return nil
	}

	return fmt.Errorf("please specify an id or name to lookup the certificate")
}
