// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/action"
)

// actionCommon implements the Configure plumbing shared by every action.
// Embed it in action implementations and access the API client via a.client
// after guarding Invoke with checkClient.
type actionCommon struct {
	client *api.Client
}

func (a *actionCommon) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.client = client
}

func (a *actionCommon) checkClient(resp *action.InvokeResponse) bool {
	if a.client == nil {
		resp.Diagnostics.AddError("Action not configured", "The API client is missing. Please report this issue to the provider developers.")
		return false
	}

	return true
}
