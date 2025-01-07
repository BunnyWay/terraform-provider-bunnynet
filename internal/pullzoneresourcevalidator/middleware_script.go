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

func MiddlewareScript() resource.ConfigValidator {
	return middlewareScriptValidator{}
}

type middlewareScriptValidator struct{}

func (v middlewareScriptValidator) Description(ctx context.Context) string {
	return "middleware_script requires the \"scripting\" routing filter"
}

func (v middlewareScriptValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v middlewareScriptValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var origin attr.Value
	req.Config.GetAttribute(ctx, path.Root("origin"), &origin)

	var routing attr.Value
	req.Config.GetAttribute(ctx, path.Root("routing"), &routing)

	if origin.IsNull() || routing.IsNull() {
		return
	}

	if origin.(types.Object).Attributes()["middleware_script"].IsUnknown() || routing.(types.Object).Attributes()["filters"].IsUnknown() {
		return
	}

	middlewareScriptId := origin.(types.Object).Attributes()["middleware_script"].(types.Int64).ValueInt64()
	filterElements := routing.(types.Object).Attributes()["filters"].(types.Set).Elements()

	var filters []string
	for _, filter := range filterElements {
		filters = append(filters, filter.(types.String).ValueString())
	}

	if middlewareScriptId > 0 && !slices.Contains(filters, "scripting") {
		resp.Diagnostics.AddError("Invalid routing.filters value", "Defining a middleware script requires the \"scripting\" routing filter")
	}
}
