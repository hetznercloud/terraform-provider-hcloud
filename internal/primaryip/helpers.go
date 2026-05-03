package primaryip

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

func AssignPrimaryIP(ctx context.Context, c *hcloud.Client, primaryIPID int64, assigneeID int64, assigneeType string) diag.Diagnostics {
	action, _, err := c.PrimaryIP.Assign(ctx, hcloud.PrimaryIPAssignOpts{
		ID:           primaryIPID,
		AssigneeID:   assigneeID,
		AssigneeType: assigneeType,
	})
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if err = c.Action.WaitFor(ctx, action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	return nil
}

func UnassignPrimaryIP(ctx context.Context, c *hcloud.Client, v int64) diag.Diagnostics {
	action, _, err := c.PrimaryIP.Unassign(ctx, v)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if err = c.Action.WaitFor(ctx, action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	return nil
}

func DeletePrimaryIP(ctx context.Context, c *hcloud.Client, p *hcloud.PrimaryIP) diag.Diagnostics {
	_, err := c.PrimaryIP.Delete(ctx, p)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	return nil
}

func CreateRandomPrimaryIP(ctx context.Context, c *hcloud.Client, server *hcloud.Server, ipType hcloud.PrimaryIPType) diag.Diagnostics {
	create, _, err := c.PrimaryIP.Create(ctx, hcloud.PrimaryIPCreateOpts{
		Name:         "primary_ip-" + strconv.Itoa(randomNumberBetween(1000000, 9999999)),
		AssigneeID:   &server.ID,
		AssigneeType: "server",
		AutoDelete:   hcloud.Ptr(true),
		Type:         ipType,
	})
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	if err = c.Action.WaitFor(ctx, create.Action); err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	return nil
}

func randomNumberBetween(low, hi int) int {
	return low + rand.Intn(hi-low) // nolint: gosec
}
