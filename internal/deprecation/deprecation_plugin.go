package deprecation

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

type DeprecationModel struct { // nolint:revive
	IsDeprecated         types.Bool   `tfsdk:"is_deprecated"`
	DeprecationAnnounced types.String `tfsdk:"deprecation_announced"`
	UnavailableAfter     types.String `tfsdk:"unavailable_after"`
}

func NewDeprecationModel(_ context.Context, in hcloud.Deprecatable) (DeprecationModel, diag.Diagnostics) {
	var data DeprecationModel
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

func DataSourceSchema(resource string) map[string]datasourceschema.Attribute {
	return map[string]datasourceschema.Attribute{
		"is_deprecated": datasourceschema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("Whether the %s is deprecated.", resource),
			Computed:            true,
		},
		"deprecation_announced": datasourceschema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Date of the %s deprecation announcement.", resource),
			Computed:            true,
		},
		"unavailable_after": datasourceschema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Date of the %s removal. After this date, the %s cannot be used anymore.", resource, resource),
			Computed:            true,
		},
	}
}

func ResourceSchema(resource string) map[string]resourceschema.Attribute {
	return map[string]resourceschema.Attribute{
		"is_deprecated": resourceschema.BoolAttribute{
			MarkdownDescription: fmt.Sprintf("Whether the %s is deprecated.", resource),
			Computed:            true,
		},
		"deprecation_announced": resourceschema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Date of the %s deprecation announcement.", resource),
			Computed:            true,
		},
		"unavailable_after": resourceschema.StringAttribute{
			MarkdownDescription: fmt.Sprintf("Date of the %s removal. After this date, the %s cannot be used anymore.", resource, resource),
			Computed:            true,
		},
	}
}
