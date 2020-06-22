package hcloud

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateCreate,
		Read:   resourceCertificateRead,
		Update: resourceCertificateUpdate,
		Delete: resourceCertificateDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
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

func resourceCertificateCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	opts := hcloud.CertificateCreateOpts{
		Name:        d.Get("name").(string),
		PrivateKey:  d.Get("private_key").(string),
		Certificate: d.Get("certificate").(string),
	}
	if labels, ok := d.GetOk("labels"); ok {
		opts.Labels = make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			opts.Labels[k] = v.(string)
		}
	}
	res, _, err := client.Certificate.Create(ctx, opts)
	if err != nil {
		return err
	}
	d.SetId(strconv.Itoa(res.ID))
	return resourceCertificateRead(d, m)
}

func resourceCertificateRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	cert, _, err := client.Certificate.Get(ctx, d.Id())
	if err != nil {
		if resourceCertificateNotFound(err, d) {
			return nil
		}
		return err
	}
	if cert == nil {
		d.SetId("")
		return nil
	}
	setCertificateSchema(d, cert)
	return nil
}

func resourceCertificateNotFound(err error, d *schema.ResourceData) bool {
	var hcloudErr hcloud.Error

	if !errors.As(err, &hcloudErr) || hcloudErr.Code != hcloud.ErrorCodeNotFound {
		return false
	}
	log.Printf("[WARN] Certificate (%s) not found, removing from state", d.Id())
	d.SetId("")
	return true
}

func setCertificateSchema(d *schema.ResourceData, cert *hcloud.Certificate) {
	d.SetId(strconv.Itoa(cert.ID))
	d.Set("name", cert.Name)
	d.Set("certificate", cert.Certificate)
	d.Set("domain_names", cert.DomainNames)
	d.Set("fingerprint", cert.Fingerprint)
	d.Set("labels", cert.Labels)
	d.Set("created", cert.Created)
	d.Set("not_valid_before", cert.NotValidBefore.Format(time.RFC3339))
	d.Set("not_valid_after", cert.NotValidAfter.Format(time.RFC3339))
}

func resourceCertificateUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	cert, _, err := client.Certificate.Get(ctx, d.Id())
	if err != nil {
		return err
	}
	if cert == nil {
		d.SetId("")
		return nil
	}

	d.Partial(true)
	if d.HasChange("name") {
		opts := hcloud.CertificateUpdateOpts{
			Name: d.Get("name").(string),
		}
		if _, _, err := client.Certificate.Update(ctx, cert, opts); err != nil {
			return err
		}
		d.SetPartial("name")
	}
	if d.HasChange("labels") {
		opts := hcloud.CertificateUpdateOpts{
			Labels: make(map[string]string),
		}
		for k, v := range d.Get("labels").(map[string]interface{}) {
			opts.Labels[k] = v.(string)
		}
		if _, _, err := client.Certificate.Update(ctx, cert, opts); err != nil {
			return err
		}
		d.SetPartial("labels")
	}
	d.Partial(false)
	return resourceCertificateRead(d, m)
}

func resourceCertificateDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	certID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid certificate id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.Certificate.Delete(ctx, &hcloud.Certificate{ID: certID}); err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// certificate has already been deleted
			return nil
		}
		return err
	}
	d.SetId("")
	return nil
}
