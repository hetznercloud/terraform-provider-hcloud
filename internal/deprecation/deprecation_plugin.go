package deprecation

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

type Model struct {
	IsDeprecated         types.Bool   `tfsdk:"is_deprecated"`
	DeprecationAnnounced types.String `tfsdk:"deprecation_announced"`
	UnavailableAfter     types.String `tfsdk:"unavailable_after"`
}

type DeprecationModel = Model //nolint:revive

func NewModel(_ context.Context, in hcloud.Deprecatable) (Model, diag.Diagnostics) {
	var data Model
	var diags diag.Diagnostics

	if in.IsDeprecated() {
		data.IsDeprecated = types.BoolValue(true)
		data.DeprecationAnnounced = types.StringValue(in.DeprecationAnnounced().Format(time.RFC3339))
		data.UnavailableAfter = types.StringValue(in.UnavailableAfter().Format(time.RFC3339))
	} else {
		data.IsDeprecated = types.BoolValue(false)

		// TODO: Stored values should be types.StringNull(), but we use an empty string
		// for backward compatibility with the SDK.
		data.DeprecationAnnounced = types.StringValue("")
		data.UnavailableAfter = types.StringValue("")
	}

	return data, diags
}

func AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_deprecated":         types.BoolType,
		"deprecation_announced": types.StringType,
		"unavailable_after":     types.StringType,
	}
}

func DataSourceSchema() map[string]datasourceschema.Attribute {
	return map[string]datasourceschema.Attribute{
		"is_deprecated": datasourceschema.BoolAttribute{
			Computed: true,
		},
		"deprecation_announced": datasourceschema.StringAttribute{
			Computed: true,
			Optional: true,
		},
		"unavailable_after": datasourceschema.StringAttribute{
			Computed: true,
			Optional: true,
		},
	}
}

func ResourceSchema() map[string]resourceschema.Attribute {
	return map[string]resourceschema.Attribute{
		"is_deprecated": resourceschema.BoolAttribute{
			Computed: true,
		},
		"deprecation_announced": resourceschema.StringAttribute{
			Computed: true,
			Optional: true,
		},
		"unavailable_after": resourceschema.StringAttribute{
			Computed: true,
			Optional: true,
		},
	}
}
