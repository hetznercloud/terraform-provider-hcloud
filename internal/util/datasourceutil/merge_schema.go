package datasourceutil

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func MergeSchema(one map[string]*schema.Schema, two map[string]*schema.Schema) map[string]*schema.Schema {
	for key, val := range two {
		one[key] = val
	}

	return one
}
