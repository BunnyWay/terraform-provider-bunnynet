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

var _ action.Action = &PullzoneCachePurgeAction{}
var _ action.ActionWithConfigure = &PullzoneCachePurgeAction{}

func NewPullzoneCachePurgeAction() action.Action {
	return &PullzoneCachePurgeAction{}
}

type PullzoneCachePurgeAction struct {
	client *api.Client
}

type PullzoneCachePurgeActionModel struct {
	Pullzone types.Int64  `tfsdk:"pullzone"`
	Tag      types.String `tfsdk:"tag"`
}

func (a *PullzoneCachePurgeAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This action allows users to purge cache for a bunny.net Pullzone.",

		Attributes: map[string]schema.Attribute{
			"pullzone": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Pullzone to purge.",
			},
			"tag": schema.StringAttribute{
				Optional:    true,
				Description: "A tag to purge. All cached objects with a matching `CDN-Tag` header will be purged.",
			},
		},
	}
}

func (a *PullzoneCachePurgeAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *PullzoneCachePurgeAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_cache_purge"
}

func (a *PullzoneCachePurgeAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data PullzoneCachePurgeActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := a.client.PullzonePurgeCache(ctx, data.Pullzone.ValueInt64(), data.Tag.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to purge cache", err.Error())
	}
}
