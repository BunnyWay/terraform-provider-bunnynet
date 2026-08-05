// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
)

var _ action.Action = &PullzonePurgeCacheAction{}
var _ action.ActionWithConfigure = &PullzonePurgeCacheAction{}

func NewPullzonePurgeCacheAction() action.Action {
	return &PullzonePurgeCacheAction{}
}

type PullzonePurgeCacheAction struct {
	actionCommon
}

type PullzonePurgeCacheActionModel struct {
	PullzoneId types.Int64  `tfsdk:"pullzone"`
	CacheTag   types.String `tfsdk:"cache_tag"`
}

func (a *PullzonePurgeCacheAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_purge_cache"
}

func (a *PullzonePurgeCacheAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This action purges the entire cache for a bunny.net pull zone. It requires Terraform 1.14 or later.",
		Attributes: map[string]schema.Attribute{
			"pullzone": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The ID of the pull zone whose cache should be purged.",
			},
			"cache_tag": schema.StringAttribute{
				Optional:    true,
				Description: "If set, only cached objects tagged with this Cache-Tag will be purged.",
			},
		},
	}
}

func (a *PullzonePurgeCacheAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data PullzonePurgeCacheActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !a.checkClient(resp) {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	err := a.client.PurgePullzoneCache(pullzoneId, data.CacheTag.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to purge pull zone cache", fmt.Sprintf("Failed to purge cache for pull zone %d: %s", pullzoneId, err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("purged cache for pull zone %d", pullzoneId))
}
