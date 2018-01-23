package hcloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
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
		return fmt.Errorf("invalid ssh key id: %v", err)
	}

	sshKey, _, err := client.SSHKey.GetByID(ctx, sshKeyID)
	if err != nil {
		return err
	}
	if sshKey == nil {
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
		return fmt.Errorf("invalid ssh key id: %v", err)
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, _, err := client.SSHKey.Update(ctx, &hcloud.SSHKey{ID: sshKeyID}, hcloud.SSHKeyUpdateOpts{
			Name: name,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceSSHKeyDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("invalid ssh key id: %v", err)
	}
	if _, err := client.SSHKey.Delete(ctx, &hcloud.SSHKey{ID: sshKeyID}); err != nil {
		return err
	}

	return nil
}
