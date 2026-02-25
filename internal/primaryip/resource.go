package primaryip

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// ResourceType is the type name of the Hetzner Cloud PrimaryIP resource.
const ResourceType = "hcloud_primary_ip"

// Resource creates a new Terraform schema for the hcloud_primary_ip resource.
func Resource() *schema.Resource {
	locationAttributes := []string{"location", "datacenter", "assignee_id"}

	return &schema.Resource{
		CreateContext: resourcePrimaryIPCreate,
		ReadContext:   resourcePrimaryIPRead,
		UpdateContext: resourcePrimaryIPUpdate,
		DeleteContext: resourcePrimaryIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"location": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: locationAttributes,
			},
			"datacenter": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Computed:     true,
				Deprecated:   "The datacenter attribute is deprecated and will be removed after 1 July 2026. Please use the location attribute instead. See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters.",
				ExactlyOneOf: locationAttributes,
			},
			"assignee_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: locationAttributes,
			},
			"assignee_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_delete": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics { // nolint:revive
					if ok, err := hcloud.ValidateResourceLabels(i.(map[string]interface{})); !ok {
						return diag.FromErr(err)
					}
					return nil
				},
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourcePrimaryIPCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	opts := hcloud.PrimaryIPCreateOpts{
		Type:         hcloud.PrimaryIPType(d.Get("type").(string)),
		AssigneeType: d.Get("assignee_type").(string),
		AutoDelete:   hcloud.Ptr(d.Get("auto_delete").(bool)),
	}
	if name, ok := d.GetOk("name"); ok {
		opts.Name = name.(string)
	}

	if assigneeID, ok := d.GetOk("assignee_id"); ok {
		opts.AssigneeID = hcloud.Ptr(util.CastInt64(assigneeID))
	} else if location, ok := d.GetOk("location"); ok {
		opts.Location = location.(string)
	} else if datacenter, ok := d.GetOk("datacenter"); ok {
		// Backward compatible datacenter argument.
		// datacenter hel1-dc2 => location hel1
		parts := strings.Split(datacenter.(string), "-")

		if len(parts) != 2 {
			return diag.Errorf("Datacenter name is not valid, expected format $LOCATION-$DATACENTER, but got: %s", datacenter.(string))
		}

		locationName := parts[0]
		opts.Location = locationName
	}

	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	res, _, err := client.PrimaryIP.Create(ctx, opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	d.SetId(util.FormatID(res.PrimaryIP.ID))
	if err = client.Action.WaitFor(ctx, res.Action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	deleteProtection := d.Get("delete_protection").(bool)
	if deleteProtection {
		if err := setProtection(ctx, client, res.PrimaryIP, deleteProtection); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	return resourcePrimaryIPRead(ctx, d, m)
}

func resourcePrimaryIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Primary IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	primaryIP, _, err := client.PrimaryIP.GetByID(ctx, id)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if primaryIP == nil {
		log.Printf("[WARN] Primary IP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setPrimaryIPSchema(d, primaryIP)
	return nil
}

func resourcePrimaryIPUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Primary IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	primaryIP := &hcloud.PrimaryIP{ID: id}

	d.Partial(true)

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, _, err := client.PrimaryIP.Update(ctx, primaryIP, hcloud.PrimaryIPUpdateOpts{
			Name: name,
		})
		if err != nil {
			if resourcePrimaryIPIsNotFound(err, d) {
				return nil
			}
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("auto_delete") {
		autoDelete := d.Get("auto_delete").(bool)
		_, _, err := client.PrimaryIP.Update(ctx, primaryIP, hcloud.PrimaryIPUpdateOpts{
			AutoDelete: hcloud.Ptr(autoDelete),
		})
		if err != nil {
			if resourcePrimaryIPIsNotFound(err, d) {
				return nil
			}
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("assignee_id") {
		serverID := util.CastInt64(d.Get("assignee_id"))
		if serverID == 0 {
			if err := UnassignPrimaryIP(ctx, client, primaryIP.ID); err != nil {
				return err
			}
		} else {
			if err := UnassignPrimaryIP(ctx, client, primaryIP.ID); err != nil {
				return err
			}
			if err := AssignPrimaryIP(ctx, client, primaryIP.ID, serverID); err != nil {
				return err
			}
		}
	}
	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.PrimaryIP.Update(ctx, primaryIP, hcloud.PrimaryIPUpdateOpts{
			Labels: &tmpLabels,
		})
		if err != nil {
			if resourcePrimaryIPIsNotFound(err, d) {
				return nil
			}
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("delete_protection") {
		deletionProtection := d.Get("delete_protection").(bool)
		if err := setProtection(ctx, client, primaryIP, deletionProtection); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	d.Partial(false)

	return resourcePrimaryIPRead(ctx, d, m)
}

func resourcePrimaryIPDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	primaryIPID, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Primary IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	if assigneeIDI, ok := d.GetOk("assignee_id"); ok && util.CastInt64(assigneeIDI) != 0 {
		assigneeID := util.CastInt64(assigneeIDI)

		if server, _, err := client.Server.GetByID(ctx, assigneeID); err == nil && server != nil {
			// The server does not have this primary ip assigned anymore, no need to try to detach it before deleting
			// Workaround for https://github.com/hashicorp/terraform/issues/35568
			if server.PublicNet.IPv4.ID == assigneeID || server.PublicNet.IPv6.ID == assigneeID {
				offAction, _, _ := client.Server.Poweroff(ctx, server)
				// if offErr != nil {
				// 	return hcloudutil.ErrorToDiag(offErr)
				// }
				if offActionErr := client.Action.WaitFor(ctx, offAction); offActionErr != nil {
					return hcloudutil.ErrorToDiag(offActionErr)
				}
				// dont catch error, because its possible that the primary IP got already unassigned on server destroy
				UnassignPrimaryIP(ctx, client, primaryIPID)

				onAction, _, _ := client.Server.Poweron(ctx, server)
				// if onErr != nil {
				// 	return hcloudutil.ErrorToDiag(onErr)
				// }
				if onActionErr := client.Action.WaitFor(ctx, onAction); onActionErr != nil {
					return hcloudutil.ErrorToDiag(onActionErr)
				}
			}
		}
	}
	err = control.Retry(2*control.DefaultRetries, func() error {
		if _, err := client.PrimaryIP.Delete(ctx, &hcloud.PrimaryIP{ID: primaryIPID}); err != nil {
			if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
				// Primary IP was already deleted
				return nil
			}
			if hcloud.IsError(err, hcloud.ErrorCodeProtected) {
				// Primary IP is delete protected
				return control.AbortRetry(err)
			}
		}
		return err
	})
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	return nil
}

func resourcePrimaryIPIsNotFound(err error, d *schema.ResourceData) bool {
	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		log.Printf("[WARN] Primary IP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setPrimaryIPSchema(d *schema.ResourceData, f *hcloud.PrimaryIP) {
	util.SetSchemaFromAttributes(d, getPrimaryIPAttributes(f))
}

func getPrimaryIPAttributes(f *hcloud.PrimaryIP) map[string]interface{} {
	res := map[string]interface{}{
		"id":                f.ID,
		"ip_address":        f.IP.String(),
		"assignee_id":       f.AssigneeID,
		"assignee_type":     f.AssigneeType,
		"name":              f.Name,
		"type":              f.Type,
		"location":          f.Location.Name,
		"labels":            f.Labels,
		"delete_protection": f.Protection.Delete,
		"auto_delete":       f.AutoDelete,
	}

	if f.Type == hcloud.PrimaryIPTypeIPv6 {
		res["ip_network"] = f.Network.String()
	}

	// Pass through datacenter name as long as it is returned from the API.
	//
	// If the attribute is not returned from the API, we never set the attribute,
	// so whatever is in the state or user config is kept.
	//
	// See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters
	//nolint:staticcheck // Backwards-compatibility
	if f.Datacenter != nil {
		//nolint:staticcheck // Backwards-compatibility
		res["datacenter"] = f.Datacenter.Name
	}

	return res
}

func setProtection(ctx context.Context, c *hcloud.Client, primaryIP *hcloud.PrimaryIP, deleteProtection bool) error {
	action, _, err := c.PrimaryIP.ChangeProtection(ctx,
		hcloud.PrimaryIPChangeProtectionOpts{
			ID:     primaryIP.ID,
			Delete: deleteProtection,
		},
	)
	if err != nil {
		return err
	}

	return c.Action.WaitFor(ctx, action)
}

func AssignPrimaryIP(ctx context.Context, c *hcloud.Client, primaryIPID int64, serverID int64) diag.Diagnostics {
	action, _, err := c.PrimaryIP.Assign(ctx, hcloud.PrimaryIPAssignOpts{
		ID:           primaryIPID,
		AssigneeID:   serverID,
		AssigneeType: "server",
	})
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if err = c.Action.WaitFor(ctx, action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	return nil
}

func UnassignPrimaryIP(ctx context.Context, c *hcloud.Client, v int64) diag.Diagnostics {
	action, _, err := c.PrimaryIP.Unassign(ctx, v)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if err = c.Action.WaitFor(ctx, action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	return nil
}

func DeletePrimaryIP(ctx context.Context, c *hcloud.Client, p *hcloud.PrimaryIP) diag.Diagnostics {
	_, err := c.PrimaryIP.Delete(ctx, p)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	return nil
}

func CreateRandomPrimaryIP(ctx context.Context, c *hcloud.Client, server *hcloud.Server, ipType hcloud.PrimaryIPType) diag.Diagnostics {
	create, _, err := c.PrimaryIP.Create(ctx, hcloud.PrimaryIPCreateOpts{
		Name:         "primary_ip-" + strconv.Itoa(randomNumberBetween(1000000, 9999999)),
		AssigneeID:   &server.ID,
		AssigneeType: "server",
		AutoDelete:   hcloud.Ptr(true),
		Type:         ipType,
	})
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if err = c.Action.WaitFor(ctx, create.Action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	return nil
}

func randomNumberBetween(low, hi int) int {
	return low + rand.Intn(hi-low) // nolint: gosec
}
