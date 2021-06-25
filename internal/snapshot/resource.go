package snapshot

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/hcclient"
)

// ResourceType is the type name of the Hetzner Cloud Snapshot resource.
const ResourceType = "hcloud_snapshot"

// Resource creates a new Terraform schema for the hcloud_floating_ip resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSnapshotCreate,
		ReadContext:   resourceSnapshotRead,
		UpdateContext: resourceSnapshotUpdate,
		DeleteContext: resourceSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	serverID := d.Get("server_id").(int)
	opts := hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Description: hcloud.String(d.Get("description").(string)),
	}

	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	res, _, err := client.Server.CreateImage(ctx, &hcloud.Server{ID: serverID}, &opts)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}

	d.SetId(strconv.Itoa(res.Image.ID))
	if res.Action != nil {
		_, errCh := client.Action.WatchProgress(ctx, res.Action)
		if err := <-errCh; err != nil {
			return hcclient.ErrorToDiag(err)
		}
	}

	return resourceSnapshotRead(ctx, d, m)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Snapshot id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	snapshot, _, err := client.Image.GetByID(ctx, id)
	if err != nil {
		return hcclient.ErrorToDiag(err)
	}
	if snapshot == nil {
		log.Printf("[WARN] Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if snapshot.Type != hcloud.ImageTypeSnapshot {
		log.Printf("[WARN] Image (%s) is not a snapshot, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setSnapshotSchema(d, snapshot)
	return nil
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Snapshot id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	image := &hcloud.Image{ID: id}

	d.Partial(true)

	if d.HasChange("description") {
		description := d.Get("description").(string)
		_, _, err := client.Image.Update(ctx, image, hcloud.ImageUpdateOpts{
			Description: hcloud.String(description),
		})
		if err != nil {
			if resourceSnapshotIsNotFound(err, d) {
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
		_, _, err := client.Image.Update(ctx, image, hcloud.ImageUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceSnapshotIsNotFound(err, d) {
				return nil
			}
			return hcclient.ErrorToDiag(err)
		}
	}
	d.Partial(false)

	return resourceSnapshotRead(ctx, d, m)
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	imageID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Snapshot id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.Image.Delete(ctx, &hcloud.Image{ID: imageID}); err != nil {
		if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
			// server has already been deleted
			return nil
		}
		return hcclient.ErrorToDiag(err)
	}

	return nil
}

func resourceSnapshotIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setSnapshotSchema(d *schema.ResourceData, s *hcloud.Image) {
	d.SetId(strconv.Itoa(s.ID))
	if s.CreatedFrom != nil {
		d.Set("server_id", s.CreatedFrom.ID)
	}
	d.Set("description", s.Description)
	d.Set("labels", s.Labels)
}
