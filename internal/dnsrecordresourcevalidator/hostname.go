package dnsrecordresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Hostname() resource.ConfigValidator {
	return hostnameValidator{}
}

type hostnameValidator struct{}

func (v hostnameValidator) Description(ctx context.Context) string {
	return "The value must not end with a dot"
}

func (v hostnameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v hostnameValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	typeAttr := path.Root("type")
	valueAttr := path.Root("value")

	var planType types.String
	req.Config.GetAttribute(ctx, typeAttr, &planType)

	var planValue types.String
	req.Config.GetAttribute(ctx, valueAttr, &planValue)

	if planValue.IsUnknown() {
		return
	}

	rType := planType.ValueString()
	value := planValue.ValueString()

	if len(value) == 0 {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(valueAttr, "Invalid attribute configuration", "Attribute cannot be empty"))
		return
	}

	valueIsHostname := rType == "CNAME" || rType == "MX" || rType == "NS" || rType == "PTR" || rType == "SRV"
	if valueIsHostname && value[len(value)-1] == '.' {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(valueAttr, "Invalid attribute configuration", v.Description(ctx)))
	}
}
