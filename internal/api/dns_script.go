package api

import (
	"context"
)

func (c *Client) UpdateDnsScript(ctx context.Context, data ComputeScript, previousData ComputeScript) (ComputeScript, error) {
	return c.UpdateComputeScript(ctx, data, previousData)
}
