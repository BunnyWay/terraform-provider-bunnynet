package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrWhitelabelNotFreeSummary = "Bunny Shield is free tier"
var ErrWhitelabelNotFreeDetail = "Whitelabel is only available for paid Bunny Shield plans."
var ErrWhitelabelDisabledSummary = "Whitelabel must be enabled"
var ErrWhitelabelDisabledDetail = "You must enable whitelabel to use custom response pages."

func Whitelabel() resource.ConfigValidator {
	return whitelabelValidator{}
}

type whitelabelValidator struct{}

func (v whitelabelValidator) Description(ctx context.Context) string {
	return "whitelabel validator"
}

func (v whitelabelValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v whitelabelValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var tier types.String
	tierAttr := path.Root("tier")
	req.Config.GetAttribute(ctx, tierAttr, &tier)

	if tier.IsUnknown() {
		return
	}

	planTypeInverted := utils.MapInvert(PlanTypeMap)
	planType, ok := planTypeInverted[tier.ValueString()]
	if !ok {
		return
	}

	isFreePlan := planType == 0

	var whitelabel types.Bool
	whitelabelAttr := path.Root("whitelabel")
	req.Config.GetAttribute(ctx, whitelabelAttr, &whitelabel)

	var whitelabelBlock types.String
	whitelabelBlockAttr := path.Root("whitelabel_block")
	req.Config.GetAttribute(ctx, whitelabelBlockAttr, &whitelabelBlock)

	var whitelabelChallenge types.String
	whitelabelChallengeAttr := path.Root("whitelabel_challenge")
	req.Config.GetAttribute(ctx, whitelabelChallengeAttr, &whitelabelChallenge)

	var whitelabelRateLimit types.String
	whitelabelRateLimitAttr := path.Root("whitelabel_rate_limit")
	req.Config.GetAttribute(ctx, whitelabelRateLimitAttr, &whitelabelRateLimit)

	if whitelabel.IsUnknown() || whitelabelBlock.IsUnknown() || whitelabelChallenge.IsUnknown() || whitelabelRateLimit.IsUnknown() {
		return
	}

	if isFreePlan {
		if whitelabel.ValueBool() {
			resp.Diagnostics.AddAttributeError(whitelabelAttr, ErrWhitelabelNotFreeSummary, ErrWhitelabelNotFreeDetail)
		}

		if v := whitelabelBlock.ValueString(); v != "" {
			resp.Diagnostics.AddAttributeError(whitelabelBlockAttr, ErrWhitelabelNotFreeSummary, ErrWhitelabelNotFreeDetail)
		}

		if v := whitelabelChallenge.ValueString(); v != "" {
			resp.Diagnostics.AddAttributeError(whitelabelChallengeAttr, ErrWhitelabelNotFreeSummary, ErrWhitelabelNotFreeDetail)
		}

		if v := whitelabelRateLimit.ValueString(); v != "" {
			resp.Diagnostics.AddAttributeError(whitelabelRateLimitAttr, ErrWhitelabelNotFreeSummary, ErrWhitelabelNotFreeDetail)
		}

		return
	}

	if !whitelabel.ValueBool() {
		if v := whitelabelBlock.ValueString(); v != "" {
			resp.Diagnostics.AddAttributeError(whitelabelAttr, ErrWhitelabelDisabledSummary, ErrWhitelabelDisabledDetail)
		}

		if v := whitelabelChallenge.ValueString(); v != "" {
			resp.Diagnostics.AddAttributeError(whitelabelAttr, ErrWhitelabelDisabledSummary, ErrWhitelabelDisabledDetail)
		}

		if v := whitelabelRateLimit.ValueString(); v != "" {
			resp.Diagnostics.AddAttributeError(whitelabelAttr, ErrWhitelabelDisabledSummary, ErrWhitelabelDisabledDetail)
		}
	}
}
