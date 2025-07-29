package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Whitelabel() resource.ConfigValidator {
	return whitelabelValidator{}
}

type whitelabelValidator struct{}

func (v whitelabelValidator) Description(ctx context.Context) string {
	return "whitelabel requires a paid \"tier\""
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

	// all but Basic
	if !ok || planType > 0 {
		return
	}

	var whitelabel types.Bool
	whitelabelAttr := path.Root("whitelabel")
	req.Config.GetAttribute(ctx, whitelabelAttr, &whitelabel)

	if whitelabel.IsUnknown() {
		return
	}

	if !whitelabel.ValueBool() {
		return
	}

	resp.Diagnostics.AddAttributeError(whitelabelAttr, "Bunny Shield is free tier", "Whitelabel is only available for paid Bunny Shield plans.")
}
