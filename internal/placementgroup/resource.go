package placementgroup

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/control"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

const ResourceType = "hcloud_placement_group"

func Resource() *schema.Resource {
	return &schema.Resource{
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
			},
			"servers": {
				Type:     schema.TypeList,
				Optional: true,
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

	setPlacementGroupSchema(d, placementGroup)
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
		if resourcePlacementGroupIsNotFound(err, d) {
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
			if resourcePlacementGroupIsNotFound(err, d) {
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
			if resourcePlacementGroupIsNotFound(err, d) {
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

	placementGroup, _, err := client.PlacementGroup.GetByID(ctx, id)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	err = control.Retry(control.DefaultRetries, func() error {
		var hcerr hcloud.Error
		_, err := client.PlacementGroup.Delete(ctx, placementGroup)
		if errors.As(err, &hcerr) {
			switch hcerr.Code {
			case hcloud.ErrorCodeNotFound:
				// placement group has already been deleted
				return nil
			case hcloud.ErrorCodeConflict, hcloud.ErrorCodeResourceInUse:
				return err
			default:
				return control.AbortRetry(err)
			}
		}
		return nil
	})
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func resourcePlacementGroupIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] placement group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setPlacementGroupSchema(d *schema.ResourceData, v *hcloud.PlacementGroup) {
	d.SetId(strconv.Itoa(v.ID))
	d.Set("name", v.Name)
	d.Set("labels", v.Labels)

	d.Set("servers", v.Servers)
	d.Set("type", v.Type)
}
