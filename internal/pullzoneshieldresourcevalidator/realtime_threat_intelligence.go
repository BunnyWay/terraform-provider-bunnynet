package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func RealtimeThreatIntelligence() resource.ConfigValidator {
	return realtimeThreatIntelligenceValidator{}
}

type realtimeThreatIntelligenceValidator struct{}

func (v realtimeThreatIntelligenceValidator) Description(ctx context.Context) string {
	return "waf.realtime_threat_intelligence requires \"tier\" = \"Advanced\""
}

func (v realtimeThreatIntelligenceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v realtimeThreatIntelligenceValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var tier types.String
	tierAttr := path.Root("tier")
	req.Config.GetAttribute(ctx, tierAttr, &tier)

	if tier.IsUnknown() || tier.ValueString() == "Advanced" {
		return
	}

	var waf types.Object
	wafAttr := path.Root("waf")
	req.Config.GetAttribute(ctx, wafAttr, &waf)

	if waf.IsUnknown() {
		return
	}

	attrs := waf.Attributes()
	rti := attrs["realtime_threat_intelligence"].(types.Bool)
	if rti.IsUnknown() {
		return
	}

	if !rti.ValueBool() {
		return
	}

	resp.Diagnostics.AddAttributeError(wafAttr, "Bunny Shield Advanced is not enabled", "Real-time Threat Intelligence is only available with Bunny Shield Advanced.")
}
