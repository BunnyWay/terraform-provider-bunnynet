package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func BotDetection() resource.ConfigValidator {
	return botDetectionValidator{}
}

type botDetectionValidator struct{}

func (v botDetectionValidator) Description(ctx context.Context) string {
	return "bot_detection requires a paid \"tier\""
}

func (v botDetectionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v botDetectionValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
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

	var botDetection types.Object
	botDetectionAttr := path.Root("bot_detection")
	req.Config.GetAttribute(ctx, botDetectionAttr, &botDetection)

	if botDetection.IsUnknown() || botDetection.IsNull() {
		return
	}

	resp.Diagnostics.AddAttributeError(botDetectionAttr, "Bunny Shield is free tier", "Bot Detection is only available for paid Bunny Shield plans.")
}
