package server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

const (
	PoweronActionType  = "hcloud_server_poweron"
	PoweroffActionType = "hcloud_server_poweroff"
	RebootActionType   = "hcloud_server_reboot"
	ResetActionType    = "hcloud_server_reset"
)

var _ action.Action = (*serverAction)(nil)
var _ action.ActionWithConfigure = (*serverAction)(nil)

type serverActionData struct {
	ServerID types.Int64 `tfsdk:"server_id"`
}

type serverActionInvoke func(ctx context.Context, client *hcloud.Client, server *hcloud.Server) (*hcloud.Action, error)

type serverAction struct {
	client              *hcloud.Client
	typeName            string
	markdownDescription string
	invoke              serverActionInvoke
}

func NewPoweronAction() action.Action {
	return &serverAction{
		typeName: PoweronActionType,
		markdownDescription: util.MarkdownDescription(`
Power on a server in Hetzner Cloud.

See the [Power on a Server documentation](https://docs.hetzner.cloud/reference/cloud#tag/server-actions/poweron_server) for more details.
`),
		invoke: func(ctx context.Context, client *hcloud.Client, server *hcloud.Server) (*hcloud.Action, error) {
			apiAction, _, err := client.Server.Poweron(ctx, server)
			return apiAction, err
		},
	}
}

func NewPoweroffAction() action.Action {
	return &serverAction{
		typeName: PoweroffActionType,
		markdownDescription: util.MarkdownDescription(`
Power off a server in Hetzner Cloud.

See the [Power off a Server documentation](https://docs.hetzner.cloud/reference/cloud#tag/server-actions/poweroff_server) for more details.
`),
		invoke: func(ctx context.Context, client *hcloud.Client, server *hcloud.Server) (*hcloud.Action, error) {
			apiAction, _, err := client.Server.Poweroff(ctx, server)
			return apiAction, err
		},
	}
}

func NewRebootAction() action.Action {
	return &serverAction{
		typeName: RebootActionType,
		markdownDescription: util.MarkdownDescription(`
Reboot a server in Hetzner Cloud.

See the [Soft-reboot a Server documentation](https://docs.hetzner.cloud/reference/cloud#tag/server-actions/reboot_server) for more details.
`),
		invoke: func(ctx context.Context, client *hcloud.Client, server *hcloud.Server) (*hcloud.Action, error) {
			apiAction, _, err := client.Server.Reboot(ctx, server)
			return apiAction, err
		},
	}
}

func NewResetAction() action.Action {
	return &serverAction{
		typeName: ResetActionType,
		markdownDescription: util.MarkdownDescription(`
Reset a server in Hetzner Cloud.

See the [Reset a Server documentation](https://docs.hetzner.cloud/reference/cloud#tag/server-actions/reset_server) for more details.
`),
		invoke: func(ctx context.Context, client *hcloud.Client, server *hcloud.Server) (*hcloud.Action, error) {
			apiAction, _, err := client.Server.Reset(ctx, server)
			return apiAction, err
		},
	}
}

func (a *serverAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = a.typeName
}

func (a *serverAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	var newDiags diag.Diagnostics

	a.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
}

func (a *serverAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		MarkdownDescription: a.markdownDescription,
		Attributes: map[string]actionschema.Attribute{
			"server_id": actionschema.Int64Attribute{
				MarkdownDescription: "ID of the server to apply the action to.",
				Required:            true,
			},
		},
	}
}

func (a *serverAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	if a.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider client is not configured. This is an issue in the provider. Please report this issue to the provider developers.",
		)
		return
	}

	var data serverActionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := &hcloud.Server{ID: data.ServerID.ValueInt64()}

	apiAction, err := a.invoke(ctx, a.client, server)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &a.client.Action, apiAction)...)
}
