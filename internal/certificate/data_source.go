package certificate

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/datasourceutil"
)

const (
	// DataSourceType is the type name of the Hetzner Cloud Certificate resource.
	DataSourceType = "hcloud_certificate"

	// DataSourceListType is the type name to receive a list of Hetzner Cloud Certificates resources.
	DataSourceListType = "hcloud_certificates"
)

// getCommonDataSchema returns a new common schema used by all certificate data sources.
func getCommonDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
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
	}
}

// DataSource creates a new Terraform schema for Hetzner Cloud Certificates.
func DataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudCertificateRead,
		Schema: datasourceutil.MergeSchema(
			getCommonDataSchema(),
			map[string]*schema.Schema{
				"with_selector": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		),
	}
}

func DataSourceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudCertificateListRead,
		Schema: map[string]*schema.Schema{
			"certificates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getCommonDataSchema(),
				},
			},
			"with_selector": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHcloudCertificateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := d.GetOk("id"); ok {
		cert, _, err := client.Certificate.GetByID(ctx, id.(int))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if cert == nil {
			return diag.Errorf("certificate not found: id: %d", id)
		}
		setCertificateSchema(d, cert)
		return nil
	}
	if name, ok := d.GetOk("name"); ok {
		cert, _, err := client.Certificate.Get(ctx, name.(string))
		if err != nil {
			return hcclient.ErrorToDiag(err)
		}
		if cert == nil {
			return diag.Errorf("certificate not found: name: %s", name)
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
			return hcclient.ErrorToDiag(err)
		}
		if len(allCertificates) == 0 {
			return hcclient.ErrorToDiag(fmt.Errorf("no Certificate found for selector %q", selector))
		}
		if len(allCertificates) > 1 {
			return hcclient.ErrorToDiag(fmt.Errorf("more than one Certificate found for selector %q", selector))
		}
		setCertificateSchema(d, allCertificates[0])
		return nil
	}

	return diag.Errorf("please specify an id or name to lookup the certificate")
}

func dataSourceHcloudCertificateListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	selector := d.Get("with_selector")
	opts := hcloud.CertificateListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: selector.(string),
		},
	}
	allCertificates, err := client.Certificate.AllWithOpts(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	ids := make([]string, len(allCertificates))
	tfCertificates := make([]map[string]interface{}, len(allCertificates))
	for i, certificate := range allCertificates {
		ids[i] = strconv.Itoa(certificate.ID)
		tfCertificates[i] = getCertificateAttributes(certificate)
	}
	d.Set("certificates", tfCertificates)
	d.SetId(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, "")))))

	return nil
}
