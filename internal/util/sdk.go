package util

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SetSchemaFromAttributes(d *schema.ResourceData, attrs map[string]any) {
	for key, value := range attrs {
		if key == "id" {
			switch v := value.(type) {
			case int64:
				d.SetId(FormatID(v))
			case int:
				d.SetId(FormatID(v))
			case string:
				d.SetId(v)
			default:
				log.Fatalf("unexpected id type '%T'", value)
			}
		} else {
			err := d.Set(key, value)
			if err != nil {
				log.Fatalf("could not set '%v' to '%s': %s", value, key, err)
			}
		}
	}
}
