// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ action.Action = &UrlCachePurgeAction{}
var _ action.ActionWithConfigure = &UrlCachePurgeAction{}

func NewUrlCachePurgeAction() action.Action {
	return &UrlCachePurgeAction{}
}

type UrlCachePurgeAction struct {
	client *api.Client
}

type UrlCachePurgeActionModel struct {
	Url       types.String `tfsdk:"url"`
	ExactPath types.Bool   `tfsdk:"exact_path"`
}

func (a *UrlCachePurgeAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This action allows users to purge cache for a specific URL in the bunny.net infrastructure.",

		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL to be purged.",
			},
			"exact_path": schema.BoolAttribute{
				Optional:    true,
				Description: "When true and the URL ends with '/', it purges only the exact path without adding a wildcard suffix. Only applies when the pull zone has IgnoreQueryStrings disabled.",
			},
		},
	}
}

func (a *UrlCachePurgeAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *UrlCachePurgeAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_url_cache_purge"
}

func (a *UrlCachePurgeAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data UrlCachePurgeActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := a.client.UrlPurgeCache(ctx, data.Url.ValueString(), data.ExactPath.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Failed to purge cache", err.Error())
	}
}
