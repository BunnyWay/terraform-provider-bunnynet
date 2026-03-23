// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/customtype"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"slices"
)

func Origin() resource.ConfigValidator {
	return originValidator{}
}

type originValidator struct{}

func (v originValidator) Description(ctx context.Context) string {
	return "Validations for origin.type"
}

func (v originValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v originValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var origin attr.Value
	req.Config.GetAttribute(ctx, path.Root("origin"), &origin)

	if origin.IsNull() {
		return
	}

	originType := origin.(types.Object).Attributes()["type"].(types.String)
	if originType.IsUnknown() {
		return
	}

	originUrl := origin.(types.Object).Attributes()["url"].(customtype.PullzoneOriginUrlValue)
	if originUrl.IsUnknown() {
		return
	}

	originStorageZone := origin.(types.Object).Attributes()["storagezone"].(types.Int64)
	if originStorageZone.IsUnknown() {
		return
	}

	originScript := origin.(types.Object).Attributes()["script"].(types.Int64)
	if originScript.IsUnknown() {
		return
	}

	originMiddlewareScript := origin.(types.Object).Attributes()["middleware_script"].(types.Int64)
	if originMiddlewareScript.IsUnknown() {
		return
	}

	originDnsPort := origin.(types.Object).Attributes()["dns_port"].(types.Int64)
	if originDnsPort.IsUnknown() {
		return
	}

	originDnsScheme := origin.(types.Object).Attributes()["dns_scheme"].(types.String)
	if originDnsScheme.IsUnknown() {
		return
	}

	var routing attr.Value
	req.Config.GetAttribute(ctx, path.Root("routing"), &routing)

	if routing.IsUnknown() {
		return
	}

	hasStorageZone := originStorageZone.ValueInt64() > 0
	hasScript := originScript.ValueInt64() > 0
	hasMiddlewareScript := originMiddlewareScript.ValueInt64() > 0
	hasScriptingRoutingFilter := false

	if !routing.IsNull() {
		routingFilters := routing.(types.Object).Attributes()["filters"]
		filterElements := routingFilters.(types.Set).Elements()

		var filters []string
		for _, filter := range filterElements {
			filters = append(filters, filter.(types.String).ValueString())
		}

		if slices.Contains(filters, "scripting") {
			hasScriptingRoutingFilter = true
		}
	}

	if originType.ValueString() != "StorageZone" && hasStorageZone {
		resp.Diagnostics.AddError("Invalid origin.storagezone value", "origin.storagezone is only applicable for StorageZone origins.")
	}

	if originType.ValueString() != "ComputeScript" && hasScript {
		resp.Diagnostics.AddError("Invalid origin.script value", "origin.script is only applicable for ComputeScript origins.")
	}

	if originType.ValueString() != "DnsAccelerate" && originDnsPort.ValueInt64() > 0 {
		resp.Diagnostics.AddError("Invalid origin.dns_port value", "origin.dns_port is only applicable for DnsAccelerate origins.")
	}

	if originType.ValueString() != "DnsAccelerate" && originDnsScheme.ValueString() != "" {
		resp.Diagnostics.AddError("Invalid origin.dns_scheme value", "origin.dns_scheme is only applicable for DnsAccelerate origins.")
	}

	if hasMiddlewareScript && !hasScriptingRoutingFilter {
		resp.Diagnostics.AddError("Invalid routing.filters value", "middleware_script requires routing.filters to contain the \"scripting\" element.")
	}

	switch originType.ValueString() {
	case "OriginUrl":
		if originUrl.IsUnknown() {
			return
		}

		if originUrl.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid origin.url value", "OriginUrl requires origin.url to be non-empty.")
		}

	case "StorageZone":
		if !originUrl.IsNull() {
			resp.Diagnostics.AddError("Invalid origin.url value", "origin.url cannot be defined for StorageZone.")
		}

		if !hasStorageZone {
			resp.Diagnostics.AddError("Invalid origin.storagezone value", "StorageZone requires origin.storagezone to be non-empty.")
		}

	case "ComputeScript":
		if originUrl.ValueString() != "https://bunnycdn.com" {
			resp.Diagnostics.AddError("Invalid origin.url value", "ComputeScript requires origin.url to be set as \"https://bunnycdn.com\".")
		}

		if !hasScript {
			resp.Diagnostics.AddError("Invalid origin.script value", "ComputeScript requires origin.script to be non-empty.")
		}

		if !hasScriptingRoutingFilter {
			resp.Diagnostics.AddError("Invalid routing.filters value", "ComputeScript requires routing.filters to contain the \"scripting\" element.")
		}

	case "DnsAccelerate":
		if !originUrl.IsNull() && !originUrl.IsUnknown() {
			resp.Diagnostics.AddError("Invalid origin.url value", "DnsAccelerate requires origin.url to be empty.")
		}

		if originDnsPort.ValueInt64() == 0 {
			resp.Diagnostics.AddError("Invalid origin.dns_port value", "DnsAccelerate requires origin.dns_port to be defined.")
		}

		if originDnsScheme.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid origin.dns_scheme value", "DnsAccelerate requires origin.dns_scheme to be defined.")
		}
	}
}
