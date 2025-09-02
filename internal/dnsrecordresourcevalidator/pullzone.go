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

func (v pullzoneIdValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	pullzoneIdAttr := path.Root("pullzone_id")
	var pullzoneId types.Int64
	request.Config.GetAttribute(ctx, pullzoneIdAttr, &pullzoneId)

	if pullzoneId.IsNull() {
		return
	}

	var rType types.String
	request.Config.GetAttribute(ctx, path.Root("type"), &rType)

	if rType.IsUnknown() || rType.IsNull() {
		return
	}

	if rType.ValueString() == "PullZone" {
		return
	}

	response.Diagnostics.AddAttributeError(pullzoneIdAttr, "Invalid attribute configuration", v.Description(ctx))
}
