package dnsrecordresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PullzoneId() resource.ConfigValidator {
	return pullzoneIdValidator{}
}

type pullzoneIdValidator struct{}

func (v pullzoneIdValidator) Description(ctx context.Context) string {
	return "pullzone_id is only available for type = PullZone"
}

func (v pullzoneIdValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v pullzoneIdValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	pullzoneIdAttr := path.Root("pullzone_id")
	typeAttr := path.Root("type")

	var planType types.String
	req.Config.GetAttribute(ctx, typeAttr, &planType)

	var planPullzoneId types.Int64
	req.Config.GetAttribute(ctx, pullzoneIdAttr, &planPullzoneId)

	if planType.IsUnknown() || planType.IsNull() {
		return
	}

	if planPullzoneId.IsUnknown() {
		return
	}

	if planType.ValueString() == "PullZone" {
		if planPullzoneId.IsNull() {
			resp.Diagnostics.AddAttributeError(pullzoneIdAttr, "Invalid attribute configuration", "pullzone_id is required")
		}

		return
	}

	if !planPullzoneId.IsNull() {
		resp.Diagnostics.AddAttributeError(pullzoneIdAttr, "Invalid attribute configuration", v.Description(ctx))
	}
}
