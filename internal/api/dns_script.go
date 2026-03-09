package api

import (
	"context"
)

func (c *Client) CreateDnsScript(ctx context.Context, data ComputeScript) (ComputeScript, error) {
	dataResult, err := c.CreateComputeScript(ctx, data)
	if err != nil {
		return dataResult, err
	}

	err = c.publishComputeScript(dataResult)
	if err != nil {
		return dataResult, err
	}

	return dataResult, nil
}

func (c *Client) UpdateDnsScript(ctx context.Context, data ComputeScript, previousData ComputeScript) (ComputeScript, error) {
	return c.UpdateComputeScript(ctx, data, previousData)
}
