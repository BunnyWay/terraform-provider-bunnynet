// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

func OriginComputeScript() resource.ConfigValidator {
	return originComputeScriptValidator{}
}

type originComputeScriptValidator struct{}

func (v originComputeScriptValidator) Description(ctx context.Context) string {
	return "Validations for origin.type = \"ComputeScript\""
}

func (v originComputeScriptValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v originComputeScriptValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var origin attr.Value
	req.Config.GetAttribute(ctx, path.Root("origin"), &origin)

	var routing attr.Value
	req.Config.GetAttribute(ctx, path.Root("routing"), &routing)

	if origin.IsNull() || routing.IsNull() {
		return
	}

	originType := origin.(types.Object).Attributes()["type"].(types.String)
	originUrl := origin.(types.Object).Attributes()["url"].(types.String)
	routingFilters := routing.(types.Object).Attributes()["filters"]

	if originType.IsUnknown() || originUrl.IsUnknown() || routingFilters.IsUnknown() {
		return
	}

	if originType.ValueString() != "ComputeScript" {
		return
	}

	if originUrl.ValueString() != "https://bunnycdn.com" {
		resp.Diagnostics.AddError("Invalid origin.url value", "ComputeScript requires origin.url to be set as \"https://bunnycdn.com\".")
	}

	filterElements := routingFilters.(types.Set).Elements()

	var filters []string
	for _, filter := range filterElements {
		filters = append(filters, filter.(types.String).ValueString())
	}

	if !slices.Contains(filters, "scripting") {
		resp.Diagnostics.AddError("Invalid routing.filters value", "ComputeScript requires routing.filters to contain the \"scripting\" element.")
	}
}
