package primaryip

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"strconv"

	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud PrimaryIP resource.
const ResourceType = "hcloud_primary_ip"

// Resource creates a new Terraform schema for the hcloud_primary_ip resource.
func Resource() *schema.Resource {
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
			"datacenter": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"assignee_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
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
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					if ok, err := hcloud.ValidateResourceLabels(i.(map[string]interface{})); !ok {
						return diag.Errorf(err.Error())
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
	assigneeID, ok1 := d.GetOk("assignee_id")
	dataCenter, ok2 := d.GetOk("datacenter")

	switch {
	case ok1 && ok2:
		return hcclient.ErrorToDiag(errors.New("assignee_id & datacenter cannot be set in the same time. " +
			"If assignee_id is set, datacenter must be left out"))
	case ok1:
		opts.AssigneeID = hcloud.Ptr(assigneeID.(int))
	case ok2:
		opts.Datacenter = dataCenter.(string)
	default:
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
		return hcclient.ErrorToDiag(err)
	}

	d.SetId(strconv.Itoa(res.PrimaryIP.ID))
	if res.Action != nil {
		_, errCh := client.Action.WatchProgress(ctx, res.Action)
		if err := <-errCh; err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	deleteProtection := d.Get("delete_protection").(bool)
	if deleteProtection {
		if err := setProtection(ctx, client, res.PrimaryIP, deleteProtection); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	return resourcePrimaryIPRead(ctx, d, m)
}

func resourcePrimaryIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Primary IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	primaryIP, _, err := client.PrimaryIP.GetByID(ctx, id)
	if err != nil {
		return hcclient.ErrorToDiag(err)
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

	id, err := strconv.Atoi(d.Id())
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
			return hcclient.ErrorToDiag(err)
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
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("assignee_id") {
		serverID := d.Get("assignee_id").(int)
		if serverID == 0 {
			if err := UnassignPrimaryIP(ctx, client, primaryIP.ID); err != nil {
				return err
			}
		} else {
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
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("delete_protection") {
		deletionProtection := d.Get("delete_protection").(bool)
		if err := setProtection(ctx, client, primaryIP, deletionProtection); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	d.Partial(false)

	return resourcePrimaryIPRead(ctx, d, m)
}

func resourcePrimaryIPDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	primaryIPID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Primary IP id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	if assigneeID, ok := d.GetOk("assignee_id"); ok && assigneeID != 0 {
		if server, _, err := client.Server.Get(ctx, strconv.Itoa(assigneeID.(int))); err == nil && server != nil {
			off, _, _ := client.Server.Poweroff(ctx, server)
			if errDiag := watchProgress(ctx, off, client); err != nil {
				return errDiag
			}
			// dont catch error, because its possible that the primary IP got already unassigned on server destroy
			UnassignPrimaryIP(ctx, client, primaryIPID)

			on, _, _ := client.Server.Poweron(ctx, server)
			if errDiag := watchProgress(ctx, on, client); err != nil {
				return errDiag
			}
		}
	}
	err = control.Retry(2*control.DefaultRetries, func() error {
		if _, err := client.PrimaryIP.Delete(ctx, &hcloud.PrimaryIP{ID: primaryIPID}); err != nil {
			if !hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
				// Primary IP was already deleted
				return err
			}
		}
		return nil
	})
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func resourcePrimaryIPIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] Primary IP (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setPrimaryIPSchema(d *schema.ResourceData, f *hcloud.PrimaryIP) {
	for key, val := range getPrimaryIPAttributes(f) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getPrimaryIPAttributes(f *hcloud.PrimaryIP) map[string]interface{} {
	res := map[string]interface{}{
		"id":                f.ID,
		"ip_address":        f.IP.String(),
		"assignee_id":       f.AssigneeID,
		"assignee_type":     f.AssigneeType,
		"name":              f.Name,
		"type":              f.Type,
		"datacenter":        f.Datacenter.Name,
		"labels":            f.Labels,
		"delete_protection": f.Protection.Delete,
		"auto_delete":       f.AutoDelete,
	}

	if f.Type == hcloud.PrimaryIPTypeIPv6 {
		res["ip_network"] = f.Network.String()
	}

	return res
}

func setProtection(ctx context.Context, c *hcloud.Client, primaryIP *hcloud.PrimaryIP, delete bool) error {
	action, _, err := c.PrimaryIP.ChangeProtection(ctx,
		hcloud.PrimaryIPChangeProtectionOpts{
			ID:     primaryIP.ID,
			Delete: delete,
		},
	)
	if err != nil {
		return err
	}

	return hcclient.WaitForAction(ctx, &c.Action, action)
}

func watchProgress(ctx context.Context, action *hcloud.Action, client *hcloud.Client) diag.Diagnostics {
	if action != nil {
		_, errCh := client.Action.WatchProgress(ctx, action)
		if err := <-errCh; err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}
	return nil
}

func AssignPrimaryIP(ctx context.Context, c *hcloud.Client, primaryIPID int, serverID int) diag.Diagnostics {
	action, _, err := c.PrimaryIP.Assign(ctx, hcloud.PrimaryIPAssignOpts{
		ID:         primaryIPID,
		AssigneeID: serverID,
	})
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return nil
}

func UnassignPrimaryIP(ctx context.Context, c *hcloud.Client, v int) diag.Diagnostics {
	action, _, err := c.PrimaryIP.Unassign(ctx, v)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return hcclient.ErrorToDiag(err)
	}
	return nil
}

func DeletePrimaryIP(ctx context.Context, c *hcloud.Client, p *hcloud.PrimaryIP) diag.Diagnostics {
	_, err := c.PrimaryIP.Delete(ctx, p)
	if err != nil {
		return hcclient.ErrorToDiag(err)
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
		return hcclient.ErrorToDiag(err)
	}

	if err := hcclient.WaitForAction(ctx, &c.Action, create.Action); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func randomNumberBetween(low, hi int) int {
	return low + rand.Intn(hi-low)
}
