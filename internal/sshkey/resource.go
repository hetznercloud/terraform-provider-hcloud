package sshkey

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
	"golang.org/x/crypto/ssh"
)

// ResourceType is the type name of the Hetzner Cloud SSH Key resource.
const ResourceType = "hcloud_ssh_key"

// Resource creates a Terraform schema for the hcloud_ssh_key resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSHKeyCreate,
		ReadContext:   resourceSSHKeyRead,
		UpdateContext: resourceSSHKeyUpdate,
		DeleteContext: resourceSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"public_key": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: resourceSSHKeyPublicKeyDiffSuppress,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceSSHKeyPublicKeyDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	fingerprint := d.Get("fingerprint").(string)
	if new != "" && fingerprint != "" {
		publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(new))
		if err != nil {
			return false
		}
		return ssh.FingerprintLegacyMD5(publicKey) == fingerprint
	}
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)
	opts := hcloud.SSHKeyCreateOpts{
		Name:      d.Get("name").(string),
		PublicKey: d.Get("public_key").(string),
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	sshKey, _, err := client.SSHKey.Create(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	d.SetId(strconv.Itoa(sshKey.ID))

	return resourceSSHKeyRead(ctx, d, m)
}

func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid SSH key id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	sshKey, _, err := client.SSHKey.GetByID(ctx, sshKeyID)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if sshKey == nil {
		log.Printf("[WARN] SSH key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setSSHKeySchema(d, sshKey)

	return nil
}

func resourceSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid SSH key id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, _, err := client.SSHKey.Update(ctx, &hcloud.SSHKey{ID: sshKeyID}, hcloud.SSHKeyUpdateOpts{
			Name: name,
		})
		if err != nil {
			if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
				log.Printf("[WARN] SSH key (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}
	if d.HasChange("labels") {
		labels := make(map[string]string)
		for k, v := range d.Get("labels").(map[string]interface{}) {
			labels[k] = v.(string)
		}
		_, _, err := client.SSHKey.Update(ctx, &hcloud.SSHKey{ID: sshKeyID}, hcloud.SSHKeyUpdateOpts{
			Labels: labels,
		})
		if err != nil {
			if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
				log.Printf("[WARN] SSH key (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	return resourceSSHKeyRead(ctx, d, m)
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid SSH key id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.SSHKey.Delete(ctx, &hcloud.SSHKey{ID: sshKeyID}); err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// SSH key has already been deleted
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func setSSHKeySchema(d *schema.ResourceData, s *hcloud.SSHKey) {
	d.SetId(strconv.Itoa(s.ID))
	d.Set("name", s.Name)
	d.Set("fingerprint", s.Fingerprint)
	d.Set("public_key", s.PublicKey)
	d.Set("labels", s.Labels)
}
