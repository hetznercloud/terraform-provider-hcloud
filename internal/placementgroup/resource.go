package placementgroup

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

const ResourceType = "hcloud_placement_group"

func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePlacementGroupCreate,
		ReadContext:   resourcePlacementGroupRead,
		UpdateContext: resourcePlacementGroupUpdate,
		DeleteContext: resourcePlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
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
			"servers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					placementGroupType := i.(string)
					switch hcloud.PlacementGroupType(placementGroupType) {
					case hcloud.PlacementGroupTypeSpread:
						return nil
					default:
						return diag.Errorf("%s is not a valid placement group type", placementGroupType)
					}
				},
			},
		},
	}
}

func resourcePlacementGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	opts := hcloud.PlacementGroupCreateOpts{
		Name: d.Get("name").(string),
		Type: hcloud.PlacementGroupType(d.Get("type").(string)),
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	res, _, err := client.PlacementGroup.Create(ctx, opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	d.SetId(strconv.Itoa(res.PlacementGroup.ID))

	if res.Action != nil {
		if err := hcclient.WaitForAction(ctx, &client.Action, res.Action); err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	return resourcePlacementGroupRead(ctx, d, m)
}

func resourcePlacementGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid placement group id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	placementGroup, _, err := client.PlacementGroup.GetByID(ctx, id)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if placementGroup == nil {
		log.Printf("[WARN] placement group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setSchema(d, placementGroup)
	return nil
}

func resourcePlacementGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid placement group id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	placementGroup, _, err := client.PlacementGroup.GetByID(ctx, id)
	if err != nil {
		if handleNotFound(err, d) {
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	d.Partial(true)

	if d.HasChange("name") {
		description := d.Get("name").(string)
		_, _, err := client.PlacementGroup.Update(ctx, placementGroup, hcloud.PlacementGroupUpdateOpts{
			Name: description,
		})
		if err != nil {
			if handleNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}

	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.PlacementGroup.Update(ctx, placementGroup, hcloud.PlacementGroupUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if handleNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}
	d.Partial(false)

	return resourcePlacementGroupRead(ctx, d, m)
}

func resourcePlacementGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid placement group id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.PlacementGroup.Delete(ctx, &hcloud.PlacementGroup{ID: id}); err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func handleNotFound(err error, d *schema.ResourceData) bool {
	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		log.Printf("[WARN] placement group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setSchema(d *schema.ResourceData, pg *hcloud.PlacementGroup) {
	for key, val := range getAttributes(pg) {
		if key == "id" {
			d.SetId(strconv.Itoa(val.(int)))
		} else {
			d.Set(key, val)
		}
	}
}

func getAttributes(pg *hcloud.PlacementGroup) map[string]interface{} {
	return map[string]interface{}{
		"id":      pg.ID,
		"name":    pg.Name,
		"labels":  pg.Labels,
		"type":    pg.Type,
		"servers": pg.Servers,
	}
}
