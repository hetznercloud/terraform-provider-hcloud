package hcloud

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func dataSourceHcloudServerType() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHcloudServerTypeRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disk": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHcloudServerTypeRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	if id, ok := data.GetOk("id"); ok {
		d, _, err := client.ServerType.GetByID(ctx, id.(int))
		if err != nil {
			return diag.FromErr(err)
		}
		if d == nil {
			return diag.Errorf("no server type found with id %d", id)
		}
		setServerTypeSchema(data, d)
		return nil
	}
	if name, ok := data.GetOk("name"); ok {
		d, _, err := client.ServerType.GetByName(ctx, name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if d == nil {
			return diag.Errorf("no server type found with name %v", name)
		}
		setServerTypeSchema(data, d)
		return nil
	}

	return diag.Errorf("please specify an id, or a name to lookup for a server type")
}

func setServerTypeSchema(data *schema.ResourceData, d *hcloud.ServerType) {
	data.SetId(strconv.Itoa(d.ID))
	data.Set("name", d.Name)
	data.Set("description", d.Description)
	data.Set("cores", d.Cores)
	data.Set("memory", d.Memory)
	data.Set("disk", d.Disk)
	data.Set("description", d.Description)
	data.Set("storage_type", d.StorageType)
	data.Set("cpu_type", d.CPUType)

}
