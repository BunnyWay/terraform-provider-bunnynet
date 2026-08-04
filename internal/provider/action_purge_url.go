// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var _ action.Action = &PurgeUrlAction{}
var _ action.ActionWithConfigure = &PurgeUrlAction{}

func NewPurgeUrlAction() action.Action {
	return &PurgeUrlAction{}
}

type PurgeUrlAction struct {
	client *api.Client
}

type PurgeUrlActionModel struct {
	Url       types.String `tfsdk:"url"`
	Async     types.Bool   `tfsdk:"async"`
	ExactPath types.Bool   `tfsdk:"exact_path"`
}

func (a *PurgeUrlAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_purge_url"
}

func (a *PurgeUrlAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This action purges a URL from the bunny.net CDN cache. It requires Terraform 1.14 or later.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^https?://.+`), "URL must start with http:// or https://"),
				},
				MarkdownDescription: "The URL to purge from the cache. Append a wildcard (`*`) to purge everything under a given path.",
			},
			"async": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the purge should be performed asynchronously: the action returns without waiting for the purge to be executed. Defaults to false.",
			},
			"exact_path": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether only the exact path should be purged, without a wildcard suffix, when the URL ends with a slash. Only applies to pull zones with query string sorting disabled. Defaults to false.",
			},
		},
	}
}

func (a *PurgeUrlAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *PurgeUrlAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data PurgeUrlActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if a.client == nil {
		resp.Diagnostics.AddError("Action not configured", "The API client is missing. Please report this issue to the provider developers.")
		return
	}

	url := data.Url.ValueString()
	err := a.client.PurgeUrl(url, data.Async.ValueBool(), data.ExactPath.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Unable to purge URL", fmt.Sprintf("Failed to purge %s: %s", url, err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("purged URL %s", url))
}
