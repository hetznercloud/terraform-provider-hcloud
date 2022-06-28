package primaryip

import (
	"context"
	"errors"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"log"
	"strconv"

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
		AutoDelete:   hcloud.Bool(d.Get("auto_delete").(bool)),
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
		opts.AssigneeID = hcloud.Int(assigneeID.(int))
		break
	case ok2:
		opts.Datacenter = dataCenter.(string)
		break
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
		if err := setProtection(ctx, client, &res.PrimaryIP, deleteProtection); err != nil {
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

	if d.HasChange("assignee_id") {
		serverID := d.Get("assignee_id").(int)
		if serverID == 0 {
			action, _, err := client.PrimaryIP.Unassign(ctx, primaryIP.ID)
			if err != nil {
				if resourcePrimaryIPIsNotFound(err, d) {
					return nil
				}
				return hcclient.ErrorToDiag(err)
			}
			if err := hcclient.WaitForAction(ctx, &client.Action, action); err != nil {
				return hcclient.ErrorToDiag(err)
			}
		} else {
			a, _, err := client.PrimaryIP.Assign(ctx, hcloud.PrimaryIPAssignOpts{
				ID:         primaryIP.ID,
				AssigneeID: serverID,
			})
			if err != nil {
				if resourcePrimaryIPIsNotFound(err, d) {
					return nil
				}
				return hcclient.ErrorToDiag(err)
			}
			if err := hcclient.WaitForAction(ctx, &client.Action, a); err != nil {
				return hcclient.ErrorToDiag(err)
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
		delete := d.Get("delete_protection").(bool)
		if err := setProtection(ctx, client, primaryIP, delete); err != nil {
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

	if assigneeID, ok := d.GetOk("assignee_id"); ok {
		shutdown, _, _ := client.Server.Shutdown(ctx, &hcloud.Server{ID: assigneeID.(int)})
		if errDiag := watchProgress(ctx, shutdown, client); err != nil {
			return errDiag
		}
		unassigned, _, _ := client.PrimaryIP.Unassign(ctx, primaryIPID)
		if errDiag := watchProgress(ctx, unassigned, client); err != nil {
			return errDiag
		}
	}
	err = control.Retry(2*control.DefaultRetries, func() error {
		if err := deletePrimaryIP(ctx, client, primaryIPID); err != nil {
			return err
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
		"ip_address":        f.IP,
		"assignee_id":       f.AssigneeID,
		"assignee_type":     f.AssigneeType,
		"name":              f.Name,
		"type":              f.Type,
		"datacenter":        f.Datacenter.Name,
		"labels":            f.Labels,
		"delete_protection": f.Protection.Delete,
		"auto_delete":       f.AutoDelete,
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

	if err := hcclient.WaitForAction(ctx, &c.Action, action); err != nil {
		return err
	}

	return nil
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

func deletePrimaryIP(ctx context.Context, client *hcloud.Client, primaryIPID int) error {
	if _, err := client.PrimaryIP.Delete(ctx, &hcloud.PrimaryIP{ID: primaryIPID}); err != nil {
		return err
	}
	return nil
}
