package server

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/primaryip"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// ResourceType is the type name of the Hetzner Cloud Server resource.
const ResourceType = "hcloud_server"

// Resource creates a Terraform schema for the hcloud_server resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
		CustomizeDiff: resourceServerCustomizeDiff,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
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
				Optional: true,
				ForceNew: true,
				ValidateFunc: func(val interface{}, key string) (i []string, errors []error) {
					image := val.(string)
					if len(image) == 0 {
						errors = append(errors, fmt.Errorf("%q must have more than 0 characters. Have you set the name instead of an ID?", key))
					}
					return
				},
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
					switch x := v.(type) {
					case string:
						return userDataHashSum(x)
					default:
						return ""
					}
				},
			},
			"ssh_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{Type: schema.TypeString,
					ValidateDiagFunc: func(i interface{}, _ cty.Path) diag.Diagnostics {
						s := i.(string)
						if len(s) == 0 {
							return diag.Errorf("Invalid ssh key passed. You need to pass a string with at least 1 character")
						}

						return nil
					}},
				ForceNew: true,
			},
			"keep_disk": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_deprecated_images": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"backup_window": {
				Type:       schema.TypeString,
				Deprecated: "You should remove this property from your terraform configuration.",
				Computed:   true,
			},
			"backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_network": {
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
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics { // nolint:revive
					if ok, err := hcloud.ValidateResourceLabels(i.(map[string]interface{})); !ok {
						return diag.Errorf(err.Error())
					}
					return nil
				},
			},
			"public_net": {
				Type:     schema.TypeSet,
				Optional: true,
				DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
					// Diff is only valid if "public_net" resource is set in
					// terraform configuration.
					_, ok := d.GetOk("public_net")
					return !ok // Negate because we do **not** want to suppress the diff.
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv4_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"ipv6_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"ipv4": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"ipv6": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"network": {
				Type:     schema.TypeSet,
				Optional: true,
				DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
					// Diff is only valid if "network" resource is set in
					// terraform configuration.
					_, ok := d.GetOk("network")
					return !ok // Negate because we do **not** want to suppress the diff.
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"alias_ips": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
							Optional: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ignore_remote_firewall_ids": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"firewall_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { // nolint:revive
					sup := d.Get("ignore_remote_firewall_ids").(bool)
					if sup && old != "" && new != "" {
						return true
					}
					return false
				},
			},
			"placement_group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"rebuild_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"shutdown_before_deletion": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"primary_disk_size": {
				Type:     schema.TypeInt,
				Computed: true,
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

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	// Get server type to select correct image (based on arch)
	serverType, _, err := c.ServerType.Get(ctx, d.Get("server_type").(string))
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if serverType == nil {
		return diag.Errorf("server type %s not found", d.Get("server_type"))
	}

	if serverType.IsDeprecated() {
		if time.Now().Before(serverType.UnavailableAfter()) {
			tflog.Warn(ctx, fmt.Sprintf("Attention: The server plan %q is deprecated and will no longer be available for order as of %s. Existing servers of that plan will continue to work as before and no action is required on your part. It is possible to migrate this server to another server plan by using the \"hcloud server change-type\" command.\n\n", serverType.Name, serverType.UnavailableAfter().Format("2006-01-02")))
		} else {
			return diag.Errorf("Attention: The server plan %q is deprecated and can no longer be ordered. Existing servers of that plan will continue to work as before and no action is required on your part. It is possible to migrate this server to another server plan by using the \"hcloud server change-type\" command.\n\n", serverType.Name)
		}
	}

	imageNameOrID := d.Get("image").(string)
	image, _, err := c.Image.GetForArchitecture(ctx, imageNameOrID, serverType.Architecture)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if image == nil {
		return diag.Errorf("image %s for architecture %s not found", imageNameOrID, serverType.Architecture)
	}

	if image.IsDeprecated() {
		if d.Get("allow_deprecated_images").(bool) {
			tflog.Warn(ctx, fmt.Sprintf("image %s is deprecated. It will continue to be available until %s", image.Name, image.Deprecated.AddDate(0, 3, 0).Format("2006-01-02")))
		} else {
			return diag.Errorf("image %s is deprecated. It will continue to be available until %s. If you want to use it, specify the allow_deprecated_images option.", image.Name, image.Deprecated.AddDate(0, 3, 0).Format("2006-01-02"))
		}
	}
	opts := hcloud.ServerCreateOpts{
		Name: d.Get("name").(string),
		ServerType: &hcloud.ServerType{
			Name: d.Get("server_type").(string),
		},
		Image:    image,
		UserData: d.Get("user_data").(string),
	}

	opts.SSHKeys, err = getSSHkeys(ctx, c, d)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if datacenter, ok := d.GetOk("datacenter"); ok {
		opts.Datacenter = &hcloud.Datacenter{Name: datacenter.(string)}
	}

	if location, ok := d.GetOk("location"); ok {
		opts.Location = &hcloud.Location{Name: location.(string)}
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	if firewallIDs, ok := d.GetOk("firewall_ids"); ok {
		for _, firewallID := range firewallIDs.(*schema.Set).List() {
			opts.Firewalls = append(opts.Firewalls, &hcloud.ServerCreateFirewall{Firewall: hcloud.Firewall{ID: firewallID.(int)}})
		}
	}

	if placementGroupID, ok := d.GetOk("placement_group_id"); ok {
		placementGroup, err := getPlacementGroup(ctx, c, placementGroupID.(int))
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}

		opts.PlacementGroup = placementGroup
	}

	if publicNet, ok := d.GetOk("public_net"); ok {
		createPublicNet := hcloud.ServerCreatePublicNet{}
		for _, publicNetBlock := range publicNet.(*schema.Set).List() {
			publicNetEntry := publicNetBlock.(map[string]interface{})
			if enableIPv4, err := toServerPublicNet[bool](publicNetEntry, "ipv4_enabled"); err == nil {
				createPublicNet.EnableIPv4 = enableIPv4
			}
			if enableIPv6, err := toServerPublicNet[bool](publicNetEntry, "ipv6_enabled"); err == nil {
				createPublicNet.EnableIPv6 = enableIPv6
			}
			if ipv4, err := toServerPublicNet[int](publicNetEntry, "ipv4"); err == nil && ipv4 != 0 {
				createPublicNet.EnableIPv4 = true
				createPublicNet.IPv4 = &hcloud.PrimaryIP{ID: ipv4}
			}
			if ipv6, err := toServerPublicNet[int](publicNetEntry, "ipv6"); err == nil && ipv6 != 0 {
				createPublicNet.EnableIPv6 = true
				createPublicNet.IPv6 = &hcloud.PrimaryIP{ID: ipv6}
			}
		}
		opts.PublicNet = &createPublicNet
		// if the server has no public net, it has to be created without starting it
		if err := onServerCreateWithoutPublicNet(&opts, d, func(opts *hcloud.ServerCreateOpts) error {
			opts.StartAfterCreate = hcloud.Ptr(false)
			return nil
		}); err != nil {
			return err
		}
	}
	res, _, err := c.Server.Create(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	d.SetId(strconv.Itoa(res.Server.ID))

	if err := hcloudutil.WaitForAction(ctx, &c.Action, res.Action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	for _, nextAction := range res.NextActions {
		if err := hcloudutil.WaitForAction(ctx, &c.Action, nextAction); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if nwSet, ok := d.GetOk("network"); ok {
		for _, item := range nwSet.(*schema.Set).List() {
			nwData := item.(map[string]interface{})
			if err := inlineAttachServerToNetwork(ctx, c, res.Server, nwData); err != nil {
				return hcloudutil.ErrorToDiag(err)
			}
		}
		// if the server was created without public net, the server is now still offline and has to be powered on after
		// network assignment
		if err := onServerCreateWithoutPublicNet(&opts, d, func(_ *hcloud.ServerCreateOpts) error {
			if err := powerOnServer(ctx, c, res.Server); err != nil {
				return err
			}

			// Workaround for network interface issue
			powerOffTwo, _, err := c.Server.Poweroff(ctx, res.Server)
			if err != nil {
				return err
			}
			if err := hcloudutil.WaitForAction(ctx, &c.Action, powerOffTwo); err != nil {
				return fmt.Errorf("power off server: %v", err)
			}
			return powerOnServer(ctx, c, res.Server)
		}); err != nil {
			return err
		}
	}

	backups := d.Get("backups").(bool)
	if err := setBackups(ctx, c, res.Server, backups); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if iso, ok := d.GetOk("iso"); ok {
		if err := setISO(ctx, c, res.Server, iso.(string)); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if rescue, ok := d.GetOk("rescue"); ok {
		if err := setRescue(ctx, c, res.Server, rescue.(string), opts.SSHKeys); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	deleteProtection := d.Get("delete_protection").(bool)
	rebuildProtection := d.Get("rebuild_protection").(bool)
	if deleteProtection || rebuildProtection {
		if err := setProtection(ctx, c, res.Server, deleteProtection, rebuildProtection); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	return resourceServerRead(ctx, d, m)
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	server, _, err := client.Server.Get(ctx, d.Id())
	if err != nil {
		if resourceServerIsNotFound(err, d) {
			return nil
		}
		return hcloudutil.ErrorToDiag(err)
	}
	if server == nil {
		d.SetId("")
		return nil
	}
	setServerSchema(d, server)

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": server.PublicNet.IPv4.IP.String(),
	})

	return nil
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hcloud.Client)

	server, _, err := c.Server.Get(ctx, d.Id())
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if server == nil {
		d.SetId("")
		return nil
	}

	d.Partial(true)
	if d.HasChange("name") {
		newName := d.Get("name")
		_, _, err := c.Server.Update(ctx, server, hcloud.ServerUpdateOpts{
			Name: newName.(string),
		})
		if err != nil {
			if resourceServerIsNotFound(err, d) {
				return nil
			}
			return hcloudutil.ErrorToDiag(err)
		}
	}
	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := c.Server.Update(ctx, server, hcloud.ServerUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceServerIsNotFound(err, d) {
				return nil
			}
			return hcloudutil.ErrorToDiag(err)
		}
	}
	if d.HasChange("server_type") {
		serverType := d.Get("server_type").(string)
		keepDisk := d.Get("keep_disk").(bool)

		if server.Status == hcloud.ServerStatusRunning {
			action, _, err := c.Server.Poweroff(ctx, server)
			if err != nil {
				return hcloudutil.ErrorToDiag(err)
			}
			if err := hcloudutil.WaitForAction(ctx, &c.Action, action); err != nil {
				return hcloudutil.ErrorToDiag(err)
			}
		}

		action, _, err := c.Server.ChangeType(ctx, server, hcloud.ServerChangeTypeOpts{
			ServerType:  &hcloud.ServerType{Name: serverType},
			UpgradeDisk: !keepDisk,
		})
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if err := hcloudutil.WaitForAction(ctx, &c.Action, action); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("backups") {
		backups := d.Get("backups").(bool)
		if err := setBackups(ctx, c, server, backups); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("iso") {
		iso := d.Get("iso").(string)
		if err := setISO(ctx, c, server, iso); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("rescue") {
		rescue := d.Get("rescue").(string)
		sshKeys, err := getSSHkeys(ctx, c, d)
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		if err := setRescue(ctx, c, server, rescue, sshKeys); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("network") {
		data := d.Get("network").(*schema.Set)
		if err := updateServerInlineNetworkAttachments(ctx, c, data, server); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("firewall_ids") {
		firewallIDs := d.Get("firewall_ids").(*schema.Set).List()
		for _, f := range server.PublicNet.Firewalls {
			found := false
			for _, i := range firewallIDs {
				fID := i.(int)
				if f.Firewall.ID == fID {
					found = true

					break
				}
			}

			if !found {
				a, _, err := c.Firewall.RemoveResources(ctx,
					&f.Firewall,
					[]hcloud.FirewallResource{
						{
							Type:   hcloud.FirewallResourceTypeServer,
							Server: &hcloud.FirewallResourceServer{ID: server.ID},
						},
					},
				)
				if err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
				err = hcloudutil.WaitForActions(ctx, &c.Action, a)
				if err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
			}
		}

		for _, i := range firewallIDs {
			fID := i.(int)
			found := false
			for _, f := range server.PublicNet.Firewalls {
				if f.Firewall.ID == fID {
					found = true

					break
				}
			}

			if !found {
				a, _, err := c.Firewall.ApplyResources(ctx,
					&hcloud.Firewall{ID: fID},
					[]hcloud.FirewallResource{
						{
							Type:   hcloud.FirewallResourceTypeServer,
							Server: &hcloud.FirewallResourceServer{ID: server.ID},
						},
					},
				)
				if err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
				err = hcloudutil.WaitForActions(ctx, &c.Action, a)
				if err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
			}
		}
	}

	if d.HasChange("public_net") {
		o, n := d.GetChange("public_net")
		if err := updatePublicNet(ctx, o, n, c, server); err != nil {
			return err
		}
	}

	if d.HasChange("placement_group_id") {
		placementGroupID := d.Get("placement_group_id").(int)
		if err := setPlacementGroup(ctx, c, server, placementGroupID); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("delete_protection") || d.HasChange("rebuild_protection") {
		deletionProtection := d.Get("delete_protection").(bool)
		rebuild := d.Get("rebuild_protection").(bool)
		if err := setProtection(ctx, c, server, deletionProtection, rebuild); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	d.Partial(false)
	return resourceServerRead(ctx, d, m)
}

func updatePublicNet(ctx context.Context, o interface{}, n interface{}, c *hcloud.Client, server *hcloud.Server) diag.Diagnostics {
	diffToRemove := o.(*schema.Set).Difference(n.(*schema.Set))
	diffToAdd := n.(*schema.Set).Difference(o.(*schema.Set))

	ipv4IDToRemove := 0
	ipv6IDToRemove := 0
	ipv4EnabledInRemoveDiff := true
	ipv6EnabledInRemoveDiff := true
	// collect ip IDs which got removed
	for _, d := range diffToRemove.List() {
		fields := d.(map[string]interface{})
		ipv4IDToRemove, ipv6IDToRemove = collectPrimaryIPIDs(fields)
		if ipv4Enabled, err := toPublicNetPrimaryIPField[bool](fields, "ipv4_enabled"); err == nil {
			ipv4EnabledInRemoveDiff = ipv4Enabled
		}
		if ipv6Enabled, err := toPublicNetPrimaryIPField[bool](fields, "ipv6_enabled"); err == nil {
			ipv6EnabledInRemoveDiff = ipv6Enabled
		}
	}

	shutdown, _, err := c.Server.Poweroff(ctx, &hcloud.Server{ID: server.ID})
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if err := hcloudutil.WaitForAction(ctx, &c.Action, shutdown); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	// if public net block is removed, auto generate primary ips & remove existing
	if diffToAdd.Len() == 0 {
		if err := publicNetRemovedDecision(ctx,
			c,
			server,
			server.PublicNet.IPv4.ID,
			ipv4IDToRemove,
			hcloud.PrimaryIPTypeIPv4); err != nil {
			return err
		}
		if err := publicNetRemovedDecision(ctx,
			c,
			server,
			server.PublicNet.IPv6.ID,
			ipv6IDToRemove,
			hcloud.PrimaryIPTypeIPv6); err != nil {
			return err
		}
	}

	// Check ip bool together with IDs to trigger the right actions
	for _, d := range diffToAdd.List() {
		fields := d.(map[string]interface{})
		ipv4IDToAdd, ipv6IDToAdd := collectPrimaryIPIDs(fields)

		if ipv4Enabled, err := toPublicNetPrimaryIPField[bool](fields, "ipv4_enabled"); err == nil {
			if err := publicNetUpdateDecision(ctx,
				c,
				ipv4Enabled,
				ipv4EnabledInRemoveDiff,
				ipv4IDToAdd,
				ipv4IDToRemove,
				server,
				server.PublicNet.IPv4.ID,
				hcloud.PrimaryIPTypeIPv4); err != nil {
				return err
			}
		}
		if ipv6Enabled, err := toPublicNetPrimaryIPField[bool](fields, "ipv6_enabled"); err == nil {
			if err := publicNetUpdateDecision(ctx,
				c, ipv6Enabled,
				ipv6EnabledInRemoveDiff,
				ipv6IDToAdd,
				ipv6IDToRemove,
				server,
				server.PublicNet.IPv6.ID,
				hcloud.PrimaryIPTypeIPv6); err != nil {
				return err
			}
		}
	}

	if err := powerOnServer(ctx, c, server); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	return nil
}

func publicNetUpdateDecision(ctx context.Context,
	c *hcloud.Client,
	ipEnabled bool,
	ipEnabledInRemoveDiff bool,
	ipID int,
	ipIDInRemoveDiff int,
	server *hcloud.Server,
	serverIPID int,
	ipType hcloud.PrimaryIPType) diag.Diagnostics {
	switch {
	// if ip set true + ip id, remove all previous assigned ipv4 + assign new
	case ipEnabled && ipID != 0:
		if serverIPID != 0 {
			// if primary ip is managed + unassigned before, this might throw an error
			if err := primaryip.UnassignPrimaryIP(ctx, c, serverIPID); err != nil {
				return err
			}
			if ipIDInRemoveDiff == 0 {
				if err := primaryip.DeletePrimaryIP(ctx, c, &hcloud.PrimaryIP{ID: serverIPID}); err != nil {
					if err := powerOnServer(ctx, c, server); err != nil {
						return hcloudutil.ErrorToDiag(err)
					}
					return err
				}
			}
		}
		if err := primaryip.AssignPrimaryIP(ctx, c, ipID, server.ID); err != nil {
			if err := powerOnServer(ctx, c, server); err != nil {
				return hcloudutil.ErrorToDiag(err)
			}
			return err
		}
	// if ip set from true -> false + no ip id, unassign + delete PrimaryIP
	case !ipEnabled && ipID == 0:
		if serverIPID != 0 {
			if err := primaryip.UnassignPrimaryIP(ctx, c, serverIPID); err != nil {
				if err := powerOnServer(ctx, c, server); err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
				return err
			}
			if ipIDInRemoveDiff == 0 {
				if err := primaryip.DeletePrimaryIP(ctx, c, &hcloud.PrimaryIP{ID: serverIPID}); err != nil {
					if err := powerOnServer(ctx, c, server); err != nil {
						return hcloudutil.ErrorToDiag(err)
					}
					return err
				}
			}
		}

	// if ip set from false -> true, create & assign auto generated primary ip
	case ipEnabled && ipID == 0:
		// unassign managed ip when id is removed
		if ipEnabledInRemoveDiff && ipIDInRemoveDiff != 0 {
			if err := primaryip.UnassignPrimaryIP(ctx, c, ipIDInRemoveDiff); err != nil {
				if err := powerOnServer(ctx, c, server); err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
				return err
			}
		}
		if !ipEnabledInRemoveDiff && ipIDInRemoveDiff == 0 ||
			ipEnabledInRemoveDiff && ipIDInRemoveDiff != 0 {
			if err := primaryip.CreateRandomPrimaryIP(ctx, c, server, ipType); err != nil {
				if err := powerOnServer(ctx, c, server); err != nil {
					return hcloudutil.ErrorToDiag(err)
				}
				return err
			}
		}

	// error on ip set from true -> false + ipv4 ID provided
	case !ipEnabled && ipID != 0:
		if err := powerOnServer(ctx, c, server); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
		return hcloudutil.ErrorToDiag(
			fmt.Errorf("this operation is not allowed: ipv4_enabled = false | ipv4 = %d", ipID))
	}
	return nil
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	serverID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid server id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	var warnings diag.Diagnostics

	if d.Get("shutdown_before_deletion").(bool) {
		// Try shutting down the server
		shutdownResult, _, err := client.Server.Shutdown(ctx, &hcloud.Server{ID: serverID})
		if err != nil {
			return hcloudutil.ErrorToDiag(err)
		}

		if err = hcloudutil.WaitForAction(ctx, &client.Action, shutdownResult); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}

		// Give the server some time to shut down
		err = control.Retry(control.DefaultRetries, func() error {
			server, _, err := client.Server.GetByID(ctx, serverID)

			// If it is not possible to get the server status, it is probably futile to retry
			if err != nil {
				return control.AbortRetry(err)
			}

			if server.Status != hcloud.ServerStatusOff {
				return fmt.Errorf("Server has not shut down yet")
			}

			// Server has shut down successfully
			return nil
		})
		if err != nil {
			// If shutting down does not work, add a warning and move on with deletion
			warnings = append(warnings, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Server id %d did not finish shutting down gracefully in time, deleting it anyways.", serverID),
			})
		}
	}

	result, _, err := client.Server.DeleteWithResult(ctx, &hcloud.Server{ID: serverID})
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	err = hcloudutil.WaitForAction(ctx, &client.Action, result.Action)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	return warnings
}

func resourceServerIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setBackups(ctx context.Context, c *hcloud.Client, server *hcloud.Server, backups bool) error {
	if server.BackupWindow != "" && !backups {
		action, _, err := c.Server.DisableBackup(ctx, server)
		if err != nil {
			return err
		}

		return hcloudutil.WaitForAction(ctx, &c.Action, action)
	}

	if server.BackupWindow == "" && backups {
		action, _, err := c.Server.EnableBackup(ctx, server, "")
		if err != nil {
			return err
		}
		return hcloudutil.WaitForAction(ctx, &c.Action, action)
	}

	return nil
}

func setISO(ctx context.Context, c *hcloud.Client, server *hcloud.Server, isoIDOrName string) error {
	isoChange := false
	if server.ISO != nil {
		isoChange = true
		action, _, err := c.Server.DetachISO(ctx, server)
		if err != nil {
			return err
		}
		if err := hcloudutil.WaitForAction(ctx, &c.Action, action); err != nil {
			return err
		}
	}
	if isoIDOrName != "" {
		isoChange = true

		iso, _, err := c.ISO.Get(ctx, isoIDOrName)
		if err != nil {
			return err
		}

		if iso == nil {
			return fmt.Errorf("ISO not found: %s", isoIDOrName)
		}

		// If ISO architecture is empty -> wildcard/unknown     --> allow
		// If ISO architecture is set and does not match server -->  deny
		if iso.Architecture != nil && *iso.Architecture != server.ServerType.Architecture {
			return errors.New("failed to attach iso: iso has a different architecture than the server")
		}

		a, _, err := c.Server.AttachISO(ctx, server, iso)
		if err != nil {
			return err
		}
		if err := hcloudutil.WaitForAction(ctx, &c.Action, a); err != nil {
			return err
		}
	}

	if isoChange {
		a, _, err := c.Server.Reset(ctx, server)
		if err != nil {
			return err
		}
		if err := hcloudutil.WaitForAction(ctx, &c.Action, a); err != nil {
			return err
		}
	}
	return nil
}

func setRescue(ctx context.Context, c *hcloud.Client, server *hcloud.Server, rescue string, sshKeys []*hcloud.SSHKey) error {
	const op = "hcloud/setRescue"

	rescueChanged := false
	if server.RescueEnabled {
		rescueChanged = true
		a, _, err := c.Server.DisableRescue(ctx, server)
		if err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
		if err := hcloudutil.WaitForAction(ctx, &c.Action, a); err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
	}
	if rescue != "" {
		rescueChanged = true
		err := control.Retry(control.DefaultRetries, func() error {
			res, _, err := c.Server.EnableRescue(ctx, server, hcloud.ServerEnableRescueOpts{
				Type:    hcloud.ServerRescueType(rescue),
				SSHKeys: sshKeys,
			})
			if err != nil {
				return err
			}
			return hcloudutil.WaitForAction(ctx, &c.Action, res.Action)
		})
		if err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
	}
	if rescueChanged {
		err := control.Retry(control.DefaultRetries*2, func() error {
			action, _, err := c.Server.Reset(ctx, server)
			if err != nil {
				return fmt.Errorf("%s: %v", op, err)
			}
			return hcloudutil.WaitForAction(ctx, &c.Action, action)
		})
		if err != nil {
			return fmt.Errorf("%s: %v", op, err)
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

func inlineAttachServerToNetwork(ctx context.Context, c *hcloud.Client, s *hcloud.Server, nwData map[string]interface{}) error {
	const op = "hcloud/inlineAttachServerToNetwork"

	nw := &hcloud.Network{ID: nwData["network_id"].(int)}
	ip := net.ParseIP(nwData["ip"].(string))

	aliasIPs := make([]net.IP, 0, nwData["alias_ips"].(*schema.Set).Len())
	for _, v := range nwData["alias_ips"].(*schema.Set).List() {
		aliasIP := net.ParseIP(v.(string))
		aliasIPs = append(aliasIPs, aliasIP)
	}
	if err := attachServerToNetwork(ctx, c, s, nw, ip, aliasIPs); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

func updateServerInlineNetworkAttachments(ctx context.Context, c *hcloud.Client, data *schema.Set, s *hcloud.Server) error {
	const op = "hcloud/updateServerInlineNetworkAttachments"

	log.Printf("[INFO] Updating inline network attachments for server %d", s.ID)

	cfgNetworks := make(map[int]map[string]interface{}, data.Len())
	for _, v := range data.List() {
		nwData := v.(map[string]interface{})
		nwID := nwData["network_id"].(int)
		cfgNetworks[nwID] = nwData
	}

	for _, n := range s.PrivateNet {
		nwData, ok := cfgNetworks[n.Network.ID]
		if !ok {
			// The server should no longer be a member of this network.
			// Detach it.
			if err := detachServerFromNetwork(ctx, c, s, n.Network); err != nil {
				return fmt.Errorf("%s: %v", op, err)
			}
			continue
		}
		// Remove the network from the cfgNetworks map. We are going to
		// handle it right now.
		delete(cfgNetworks, n.Network.ID)

		if nwData["ip"].(string) != n.IP.String() {
			// IP changed. Our API provides now way to change this. So we
			// need to detach and re-attach. Alias IPs are updated, too. This
			// saves us from the next step.
			if err := detachServerFromNetwork(ctx, c, s, n.Network); err != nil {
				return fmt.Errorf("%s: %v", op, err)
			}
			if err := inlineAttachServerToNetwork(ctx, c, s, nwData); err != nil {
				return fmt.Errorf("%s: %v", op, err)
			}
			continue
		}
		cfgAliasIPs := nwData["alias_ips"].(*schema.Set)
		curAliasIPs := newIPSet(cfgAliasIPs.F, n.Aliases)
		if !cfgAliasIPs.Equal(curAliasIPs) {
			if err := updateServerAliasIPs(ctx, c, s, n.Network, cfgAliasIPs); err != nil {
				return fmt.Errorf("%s: %v", op, err)
			}
			continue
		}
	}

	// Whatever remains in cfgNetworks now is a newly added network. We attach
	// the server to it.
	for _, nwData := range cfgNetworks {
		if err := inlineAttachServerToNetwork(ctx, c, s, nwData); err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
	}

	return nil
}

func newIPSet(f schema.SchemaSetFunc, ips []net.IP) *schema.Set {
	ss := make([]interface{}, len(ips))
	for i, ip := range ips {
		ss[i] = ip.String()
	}
	return schema.NewSet(f, ss)
}

func setServerSchema(d *schema.ResourceData, s *hcloud.Server) {
	for key, val := range getServerAttributes(d, s) {
		switch key {
		case "id":
			d.SetId(strconv.Itoa(val.(int)))
		default:
			err := d.Set(key, val)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func getServerAttributes(d *schema.ResourceData, s *hcloud.Server) map[string]interface{} {
	firewallIDs := make([]int, len(s.PublicNet.Firewalls))
	for i, firewall := range s.PublicNet.Firewalls {
		firewallIDs[i] = firewall.Firewall.ID
	}

	res := map[string]interface{}{
		"id":                 s.ID,
		"name":               s.Name,
		"datacenter":         s.Datacenter.Name,
		"location":           s.Datacenter.Location.Name,
		"status":             s.Status,
		"server_type":        s.ServerType.Name,
		"ipv6_network":       s.PublicNet.IPv6.Network.String(),
		"backup_window":      s.BackupWindow,
		"backups":            s.BackupWindow != "",
		"labels":             s.Labels,
		"delete_protection":  s.Protection.Delete,
		"rebuild_protection": s.Protection.Rebuild,
		"firewall_ids":       firewallIDs,
		"primary_disk_size":  s.PrimaryDiskSize,
	}
	if s.PublicNet.IPv4.IsUnspecified() {
		res["ipv4_address"] = nil
	} else {
		res["ipv4_address"] = s.PublicNet.IPv4.IP.String()
	}

	if len(s.PublicNet.IPv6.IP) == 0 {
		// No IPv6 Primary IP assigned
		res["ipv6_address"] = nil
	} else {
		// Set first IP in assigned subnet range
		res["ipv6_address"] = s.PublicNet.IPv6.IP.String() + "1"
	}

	if s.Image != nil {
		if s.Image.Name != "" && strconv.Itoa(s.Image.ID) != d.Get("image") {
			// Only use the image name if the image is official (Name != "")
			// AND the user did not explicitly specify the image id
			res["image"] = s.Image.Name
		} else {
			res["image"] = fmt.Sprintf("%d", s.Image.ID)
		}
	}

	// Only write the networks to the resource data if it already contains
	// such an entry. This avoids setting the "network" property which is not
	// marked as "computed" if the user uses the "server_network_subnet"
	// resource. Setting the "network" property as computed is not possible
	// because this would lead to loosing any updates.
	//
	// The easiest would be to use schema.ComputedWhen but this is marked
	// as currently not working.
	if _, ok := d.GetOk("network"); ok {
		res["network"] = networkToTerraformNetworks(s.PrivateNet)
	}

	if s.PlacementGroup != nil {
		res["placement_group_id"] = s.PlacementGroup.ID
	} else {
		res["placement_group_id"] = nil
	}

	return res
}

func networkToTerraformNetworks(privateNetworks []hcloud.ServerPrivateNet) []map[string]interface{} {
	tfPrivateNetworks := make([]map[string]interface{}, len(privateNetworks))
	for i, privateNetwork := range privateNetworks {
		tfPrivateNetwork := make(map[string]interface{})
		tfPrivateNetwork["network_id"] = privateNetwork.Network.ID
		tfPrivateNetwork["ip"] = privateNetwork.IP.String()
		tfPrivateNetwork["mac_address"] = privateNetwork.MACAddress
		aliasIPs := make([]string, len(privateNetwork.Aliases))
		for in, ip := range privateNetwork.Aliases {
			aliasIPs[in] = ip.String()
		}
		tfPrivateNetwork["alias_ips"] = aliasIPs
		tfPrivateNetworks[i] = tfPrivateNetwork
	}
	return tfPrivateNetworks
}

func getPlacementGroup(ctx context.Context, c *hcloud.Client, id int) (*hcloud.PlacementGroup, error) {
	placementGroup, _, err := c.PlacementGroup.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if placementGroup == nil {
		return nil, fmt.Errorf("placement group not found: %d", id)
	}

	return placementGroup, nil
}

func setPlacementGroup(ctx context.Context, c *hcloud.Client, server *hcloud.Server, id int) error {
	if server.PlacementGroup != nil {
		action, _, err := c.Server.RemoveFromPlacementGroup(ctx, server)
		if err != nil {
			return err
		}
		if err := hcloudutil.WaitForAction(ctx, &c.Action, action); err != nil {
			return err
		}
	}

	if id != 0 {
		placementGroup, err := getPlacementGroup(ctx, c, id)
		if err != nil {
			return err
		}

		action, _, err := c.Server.AddToPlacementGroup(ctx, server, placementGroup)
		if err != nil {
			return err
		}
		if err := hcloudutil.WaitForAction(ctx, &c.Action, action); err != nil {
			return err
		}
	}

	return nil
}

func setProtection(ctx context.Context, c *hcloud.Client, server *hcloud.Server, delete bool, rebuild bool) error {
	action, _, err := c.Server.ChangeProtection(ctx, server,
		hcloud.ServerChangeProtectionOpts{
			Delete:  &delete,
			Rebuild: &rebuild,
		},
	)
	if err != nil {
		return err
	}

	return hcloudutil.WaitForAction(ctx, &c.Action, action)
}

func toServerPublicNet[V int | bool](field map[string]interface{}, key string) (V, error) {
	var op = "toServerPublicNet"
	var valType V
	if valType, ok := field[key].(V); ok {
		return valType, nil
	}
	return valType, fmt.Errorf("%s: unable to apply value to public_net values", op)
}

func collectPrimaryIPIDs(primaryIPList map[string]interface{}) (int, int) {
	var IPv4ID = 0
	var IPv6ID = 0
	if id, err := toPublicNetPrimaryIPField[int](primaryIPList, "ipv4"); id != 0 && err == nil {
		IPv4ID = id
	}
	if id, err := toPublicNetPrimaryIPField[int](primaryIPList, "ipv6"); id != 0 && err == nil {
		IPv6ID = id
	}
	return IPv4ID, IPv6ID
}

func toPublicNetPrimaryIPField[V int | bool](field map[string]interface{}, key string) (V, error) {
	var op = "toPublicNetPrimaryIPField"
	var fieldValue V
	if fieldValue, ok := field[key].(V); ok {
		return fieldValue, nil
	}
	return fieldValue, fmt.Errorf("%s: field does not contain ID", op)
}

func onServerCreateWithoutPublicNet(opts *hcloud.ServerCreateOpts, d *schema.ResourceData, fn func(opts *hcloud.ServerCreateOpts) error) diag.Diagnostics {
	if _, ok := d.GetOk("network"); ok && opts.PublicNet != nil {
		if !opts.PublicNet.EnableIPv6 && !opts.PublicNet.EnableIPv4 {
			if err := fn(opts); err != nil {
				return hcloudutil.ErrorToDiag(err)
			}
		}
		return nil
	}
	return nil
}

func powerOnServer(ctx context.Context, c *hcloud.Client, server *hcloud.Server) error {
	err := control.Retry(control.DefaultRetries, func() error {
		powerOn, _, err := c.Server.Poweron(ctx, server)
		if err != nil {
			return err
		}

		return hcloudutil.WaitForAction(ctx, &c.Action, powerOn)
	})
	if err != nil {
		return err
	}
	return nil
}

func publicNetRemovedDecision(ctx context.Context,
	c *hcloud.Client,
	server *hcloud.Server,
	serverIPID int,
	ipIDToRemove int,
	ipType hcloud.PrimaryIPType) diag.Diagnostics {
	if server.PublicNet.IPv4.ID != 0 && ipIDToRemove != 0 {
		if err := primaryip.UnassignPrimaryIP(ctx, c, serverIPID); err != nil {
			return err
		}
		if err := primaryip.CreateRandomPrimaryIP(ctx, c, server, ipType); err != nil {
			return err
		}
	}
	return nil
}

func resourceServerCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	return validateUniqueNetworkIDs(d)
}

func validateUniqueNetworkIDs(d *schema.ResourceDiff) error {
	// Validate that every network set element uses unique network id
	if n, ok := d.GetOkExists("network"); ok {
		networks, ok := n.(*schema.Set)
		if !ok {
			return fmt.Errorf("network has unexpected type: %T", n)
		}

		uniqueNetworkIDs := map[int]bool{}

		for _, networkI := range networks.List() {
			network, ok := networkI.(map[string]interface{})
			if !ok {
				return fmt.Errorf("network item has unexpected type: %T", networkI)
			}

			networkID, ok := network["network_id"]
			if !ok {
				continue
			}
			if networkID == 0 {
				// ID is 0 if Network will be created in same apply, we are unable to reliably detect if the
				// "to-be-created" networks are the same.
				// See https://github.com/hetznercloud/terraform-provider-hcloud/issues/899

				// We should revisit this after moving to the Plugin Framework,
				// as it allows more introspection for the plan/changes made.
				// See https://github.com/hetznercloud/terraform-provider-hcloud/issues/752
				continue
			}

			id, ok := networkID.(int)
			if !ok {
				return fmt.Errorf("network id has unexpected type: %T", networkID)
			}

			if uniqueNetworkIDs[id] {
				return fmt.Errorf("server is only allowed to be attached to each network once: %d", networkID)
			}

			uniqueNetworkIDs[id] = true
		}
	}

	return nil
}
