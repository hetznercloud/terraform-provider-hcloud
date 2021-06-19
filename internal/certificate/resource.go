package certificate

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

const (
	// ResourceType is the type name of the Hetzner Cloud resource for
	// certificates of type uploaded.
	ResourceType = "hcloud_certificate"

	// UploadedResourceType is the name of the Hetzner Cloud uploaded
	// Certificate resource. Resources of this type behave identical to
	// the hcloud_certificate resource.
	UploadedResourceType = "hcloud_uploaded_certificate"

	// ManagedResourceType is the name of the Hetzner Cloud managed Certificate
	// resource.
	ManagedResourceType = "hcloud_managed_certificate"
)

// UploadedResource creates a new Terraform schema for Hetzner Cloud Certificate
// resources.
func UploadedResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createUploadedResource,
		ReadContext:   readResource,
		UpdateContext: updateResource,
		DeleteContext: deleteResource,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(_, certOld, certNew string, d *schema.ResourceData) bool {
					res, err := EqualCert(certOld, certNew)
					if err != nil {
						log.Printf("[ERROR] compare certificates for equality: %v", err)
						return false
					}
					return res
				},
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

// ManagedResource creates a new Terraform schema for the Hetzner Cloud managed
// Certificate resource.
func ManagedResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createManagedResource,
		ReadContext:   readResource,
		UpdateContext: updateResource,
		DeleteContext: deleteResource,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_names": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
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

func createUploadedResource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

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
		return hcclient.ErrorToDiag(err)
	}
	d.SetId(strconv.Itoa(res.ID))
	return readResource(ctx, d, m)
}

func createManagedResource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	opts := hcloud.CertificateCreateOpts{
		Name: d.Get("name").(string),
		Type: hcloud.CertificateTypeManaged,
	}

	domainNameSet := d.Get("domain_names").(*schema.Set)
	opts.DomainNames = make([]string, domainNameSet.Len())
	for i, n := range domainNameSet.List() {
		opts.DomainNames[i] = n.(string)
	}

	if labels, ok := d.GetOk("labels"); ok {
		opts.Labels = make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			opts.Labels[k] = v.(string)
		}
	}

	res, _, err := c.Certificate.CreateCertificate(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, res.Action); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	d.SetId(strconv.Itoa(res.Certificate.ID))

	return readResource(ctx, d, m)
}

func readResource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	cert, _, err := client.Certificate.Get(ctx, d.Id())
	if err != nil {
		if resourceCertificateNotFound(err, d) {
			return nil
		}
		return hcclient.ErrorToDiag(err)
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
	d.Set("type", cert.Type)
	d.Set("certificate", cert.Certificate)
	d.Set("domain_names", cert.DomainNames)
	d.Set("fingerprint", cert.Fingerprint)
	d.Set("labels", cert.Labels)
	d.Set("created", cert.Created.String())
	d.Set("not_valid_before", cert.NotValidBefore.Format(time.RFC3339))
	d.Set("not_valid_after", cert.NotValidAfter.Format(time.RFC3339))
}

func updateResource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	cert, _, err := client.Certificate.Get(ctx, d.Id())
	if err != nil {
		return hcclient.ErrorToDiag(err)
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
			return hcclient.ErrorToDiag(err)
		}
	}
	if d.HasChange("labels") {
		opts := hcloud.CertificateUpdateOpts{
			Labels: make(map[string]string),
		}
		for k, v := range d.Get("labels").(map[string]interface{}) {
			opts.Labels[k] = v.(string)
		}
		if _, _, err := client.Certificate.Update(ctx, cert, opts); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	d.Partial(false)
	return readResource(ctx, d, m)
}

func deleteResource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

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
		return hcclient.ErrorToDiag(err)
	}
	d.SetId("")
	return nil
}
