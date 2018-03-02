package hcloud

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"server_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"datacenter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"user_data": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: userDataDiffSuppress,
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						return userDataHashSum(v.(string))
					default:
						return ""
					}
				},
			},
			"ssh_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"keep_disk": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"backup_window": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iso": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rescue": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func userDataHashSum(userData string) string {
	sum := sha1.Sum([]byte(userData))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func userDataDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	userData := d.Get(k).(string)
	if new != "" && userData != "" {
		if _, err := base64.StdEncoding.DecodeString(old); err != nil {
			return userDataHashSum(old) == new
		}
	}
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	opts := hcloud.ServerCreateOpts{
		Name: d.Get("name").(string),
		ServerType: &hcloud.ServerType{
			Name: d.Get("server_type").(string),
		},
		Image: &hcloud.Image{
			Name: d.Get("image").(string),
		},
		UserData: d.Get("user_data").(string),
	}

	var err error
	opts.SSHKeys, err = getSSHkeys(ctx, client, d)
	if err != nil {
		return err
	}

	if datacenter, ok := d.GetOk("datacenter"); ok {
		opts.Datacenter = &hcloud.Datacenter{Name: datacenter.(string)}
	}

	if location, ok := d.GetOk("location"); ok {
		opts.Location = &hcloud.Location{Name: location.(string)}
	}

	res, _, err := client.Server.Create(ctx, opts)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(res.Server.ID))
	_, errCh := client.Action.WatchProgress(ctx, res.Action)
	if err := <-errCh; err != nil {
		return err
	}

	backupWindow := d.Get("backup_window").(string)
	if backupWindow != "" {
		if err := setBackupWindow(ctx, client, res.Server, backupWindow); err != nil {
			return err
		}
	}

	if iso, ok := d.GetOk("iso"); ok {
		if err := setISO(ctx, client, res.Server, iso.(string)); err != nil {
			return err
		}
	}

	if rescue, ok := d.GetOk("rescue"); ok {
		if err := setRescue(ctx, client, res.Server, rescue.(string), opts.SSHKeys); err != nil {
			return err
		}
	}

	return resourceServerRead(d, m)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, _, err := client.Server.Get(ctx, d.Id())
	if err != nil {
		return err
	}
	if server == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", server.Name)
	d.Set("datacenter", server.Datacenter.Name)
	d.Set("location", server.Datacenter.Location.Name)
	d.Set("status", server.Status)
	d.Set("server_type", server.ServerType.Name)
	d.Set("ipv4_address", server.PublicNet.IPv4.IP.String())
	d.Set("ipv6_address", server.PublicNet.IPv6.IP.String())
	d.Set("backup_window", server.BackupWindow)
	if server.Image != nil {
		if server.Image.Name != "" {
			d.Set("image", server.Image.Name)
		} else {
			d.Set("image", server.Image.ID)
		}
	}

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": server.PublicNet.IPv4.IP.String(),
	})

	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, _, err := client.Server.Get(ctx, d.Id())
	if err != nil {
		return err
	}
	if server == nil {
		d.SetId("")
		return nil
	}

	d.Partial(true)
	if d.HasChange("name") {
		newName := d.Get("name")
		_, _, err := client.Server.Update(ctx, server, hcloud.ServerUpdateOpts{
			Name: newName.(string),
		})
		if err != nil {
			return err
		}
		d.SetPartial("name")
	}

	if d.HasChange("server_type") {
		serverType := d.Get("server_type").(string)
		keepDisk := d.Get("keep_disk").(bool)

		if server.Status == hcloud.ServerStatusRunning {
			action, _, err := client.Server.Poweroff(ctx, server)
			if err != nil {
				return err
			}
			_, errCh := client.Action.WatchProgress(ctx, action)
			if err := <-errCh; err != nil {
				return err
			}
		}

		action, _, err := client.Server.ChangeType(ctx, server, hcloud.ServerChangeTypeOpts{
			ServerType:  &hcloud.ServerType{Name: serverType},
			UpgradeDisk: !keepDisk,
		})
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
		d.SetPartial("server_type")
	}

	if d.HasChange("backup_window") {
		backupWindow := d.Get("backup_window").(string)
		if err := setBackupWindow(ctx, client, server, backupWindow); err != nil {
			return err
		}
		d.SetPartial("backup_window")
	}

	if d.HasChange("iso") {
		iso := d.Get("iso").(string)
		if err := setISO(ctx, client, server, iso); err != nil {
			return err
		}
		d.SetPartial("iso")
	}

	if d.HasChange("rescue") {
		rescue := d.Get("rescue").(string)
		sshKeys, err := getSSHkeys(ctx, client, d)
		if err != nil {
			return err
		}
		if err := setRescue(ctx, client, server, rescue, sshKeys); err != nil {
			return err
		}
	}

	d.Partial(false)
	return nil
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	serverID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("invalid server id: %v", err)
	}
	if _, err := client.Server.Delete(ctx, &hcloud.Server{ID: serverID}); err != nil {
		return err
	}

	return nil
}

func setBackupWindow(ctx context.Context, client *hcloud.Client, server *hcloud.Server, backupWindow string) error {
	if backupWindow == "" {
		action, _, err := client.Server.DisableBackup(ctx, server)
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
		return nil
	}

	action, _, err := client.Server.EnableBackup(ctx, server, backupWindow)
	if err != nil {
		return err
	}
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

func setISO(ctx context.Context, client *hcloud.Client, server *hcloud.Server, iso string) error {
	isoChange := false
	if server.ISO != nil {
		isoChange = true
		action, _, err := client.Server.DetachISO(ctx, server)
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
	}
	if iso != "" {
		isoChange = true
		action, _, err := client.Server.AttachISO(ctx, server, &hcloud.ISO{Name: iso})
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
	}

	if isoChange {
		action, _, err := client.Server.Reset(ctx, server)
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}

func setRescue(ctx context.Context, client *hcloud.Client, server *hcloud.Server, rescue string, sshKeys []*hcloud.SSHKey) error {
	rescueChanged := false
	if server.RescueEnabled {
		rescueChanged = true
		action, _, err := client.Server.DisableRescue(ctx, server)
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
	}
	if rescue != "" {
		rescueChanged = true
		res, _, err := client.Server.EnableRescue(ctx, server, hcloud.ServerEnableRescueOpts{
			Type:    hcloud.ServerRescueType(rescue),
			SSHKeys: sshKeys,
		})
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, res.Action)
		if err := <-errCh; err != nil {
			return err
		}
	}
	if rescueChanged {
		action, _, err := client.Server.Reset(ctx, server)
		if err != nil {
			return err
		}
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}

func getSSHkeys(ctx context.Context, client *hcloud.Client, d *schema.ResourceData) (sshKeys []*hcloud.SSHKey, err error) {
	for _, sshKeyValue := range d.Get("ssh_keys").([]interface{}) {
		sshKeyIDOrName := sshKeyValue.(string)
		var sshKey *hcloud.SSHKey
		sshKey, _, err = client.SSHKey.Get(ctx, sshKeyIDOrName)
		if err != nil {
			return
		}
		if sshKey == nil {
			err = fmt.Errorf("SSH key not found: %s", sshKeyIDOrName)
			return
		}
		sshKeys = append(sshKeys, sshKey)
	}
	return
}
