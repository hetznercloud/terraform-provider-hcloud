package hcloud

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"golang.org/x/crypto/ssh"
)

func resourceSSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceSSHKeyCreate,
		Read:   resourceSSHKeyRead,
		Update: resourceSSHKeyUpdate,
		Delete: resourceSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceSSHKeyCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	sshKey, _, err := client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      d.Get("name").(string),
		PublicKey: d.Get("public_key").(string),
	})
	if err != nil {
		return err
	}
	d.SetId(strconv.Itoa(sshKey.ID))

	return resourceSSHKeyRead(d, m)
}

func resourceSSHKeyRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid SSH key id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	sshKey, _, err := client.SSHKey.GetByID(ctx, sshKeyID)
	if err != nil {
		return err
	}
	if sshKey == nil {
		log.Printf("[WARN] SSH key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", sshKey.Name)
	d.Set("fingerprint", sshKey.Fingerprint)

	return nil
}

func resourceSSHKeyUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

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
			return err
		}
		d.SetPartial("name")
	}

	return resourceSSHKeyRead(d, m)
}

func resourceSSHKeyDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

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
		return err
	}

	return nil
}
