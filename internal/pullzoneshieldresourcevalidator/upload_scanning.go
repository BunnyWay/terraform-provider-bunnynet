package pullzoneshieldresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UploadScanning() resource.ConfigValidator {
	return uploadScanningValidator{}
}

type uploadScanningValidator struct{}

func (v uploadScanningValidator) Description(ctx context.Context) string {
	return "upload_scanning_antivirus requires a paid \"tier\""
}

func (v uploadScanningValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uploadScanningValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
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

	var uploadScanningAntivirus types.String
	attr := path.Root("upload_scanning_antivirus")
	req.Config.GetAttribute(ctx, attr, &uploadScanningAntivirus)

	if uploadScanningAntivirus.IsUnknown() || uploadScanningAntivirus.IsNull() {
		return
	}

	if uploadScanningAntivirus.ValueString() == "Disable" {
		return
	}

	resp.Diagnostics.AddAttributeError(attr, "Bunny Shield is free tier", "Antivirus Upload Scanning is only available for paid Bunny Shield plans.")
}
