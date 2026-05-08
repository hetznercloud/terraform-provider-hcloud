package server

import (
	"context"
	"fmt"

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
	RebuildActionType  = "hcloud_server_rebuild"
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

// Rebuild action

var _ action.Action = (*rebuildAction)(nil)
var _ action.ActionWithConfigure = (*rebuildAction)(nil)

type rebuildActionData struct {
	ServerID types.Int64  `tfsdk:"server_id"`
	Image    types.String `tfsdk:"image"`
	UserData types.String `tfsdk:"user_data"`
}

type rebuildAction struct {
	client *hcloud.Client
}

func NewRebuildAction() action.Action {
	return &rebuildAction{}
}

func (a *rebuildAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = RebuildActionType
}

func (a *rebuildAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	var newDiags diag.Diagnostics

	a.client, newDiags = hcloudutil.ConfigureClient(req.ProviderData)
	resp.Diagnostics.Append(newDiags...)
}

func (a *rebuildAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionschema.Schema{
		MarkdownDescription: util.MarkdownDescription(`
Rebuild a server in Hetzner Cloud.

~> **Warning:** Rebuilding a server will destroy all data on the server's disk.
Make sure you have backups before proceeding.

See the [Rebuild a Server documentation](https://docs.hetzner.cloud/reference/cloud#tag/server-actions/rebuild_server) for more details.
`),
		Attributes: map[string]actionschema.Attribute{
			"server_id": actionschema.Int64Attribute{
				MarkdownDescription: "ID of the server to rebuild.",
				Required:            true,
			},
			"image": actionschema.StringAttribute{
				MarkdownDescription: "Name or ID of the image to rebuild the server from.",
				Required:            true,
			},
			"user_data": actionschema.StringAttribute{
				MarkdownDescription: "Cloud-Init user data to use when rebuilding the server.",
				Optional:            true,
			},
		},
	}
}

func (a *rebuildAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	if a.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider client is not configured. This is an issue in the provider. Please report this issue to the provider developers.",
		)
		return
	}

	var data rebuildActionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, _, err := a.client.Server.GetByID(ctx, data.ServerID.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
	if server == nil {
		resp.Diagnostics.AddError(
			"Server not found",
			fmt.Sprintf("Server with ID %d not found.", data.ServerID.ValueInt64()),
		)
		return
	}

	// Look up the image by name or ID, scoped to the server's architecture.
	image, _, err := a.client.Image.GetForArchitecture(ctx, data.Image.ValueString(), server.ServerType.Architecture)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}
	if image == nil {
		resp.Diagnostics.AddError(
			"Image not found",
			fmt.Sprintf(
				"Image %q for architecture %s not found.",
				data.Image.ValueString(), server.ServerType.Architecture,
			),
		)
		return
	}

	opts := hcloud.ServerRebuildOpts{
		Image: image,
	}

	if !data.UserData.IsNull() && !data.UserData.IsUnknown() {
		ud := data.UserData.ValueString()
		opts.UserData = &ud
	}

	apiAction, _, err := a.client.Server.Rebuild(ctx, server, opts)
	if err != nil {
		resp.Diagnostics.Append(hcloudutil.APIErrorDiagnostics(err)...)
		return
	}

	resp.Diagnostics.Append(hcloudutil.SettleActions(ctx, &a.client.Action, apiAction)...)
}
