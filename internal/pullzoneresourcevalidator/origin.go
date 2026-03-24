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
	"net/url"
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

	var routing attr.Value
	req.Config.GetAttribute(ctx, path.Root("routing"), &routing)

	if routing.IsUnknown() {
		return
	}

	originPath := path.Root("origin")
	originTypePath := originPath.AtName("type")
	originAttr := origin.(types.Object).Attributes()

	originType := originAttr["type"].(types.String)
	if originType.IsUnknown() {
		return
	}

	originUrl := originAttr["url"].(customtype.PullzoneOriginUrlValue)
	if originUrl.IsUnknown() {
		return
	}

	originHostHeader := originAttr["host_header"].(types.String)
	if originHostHeader.IsUnknown() {
		return
	}

	originStorageZone := originAttr["storagezone"].(types.Int64)
	if originStorageZone.IsUnknown() {
		return
	}

	originScript := originAttr["script"].(types.Int64)
	originMiddlewareScript := originAttr["middleware_script"].(types.Int64)
	if originMiddlewareScript.IsUnknown() {
		return
	}

	originContainerAppId := originAttr["container_app_id"].(types.String)
	if originContainerAppId.IsUnknown() {
		return
	}

	originContainerEndpointId := originAttr["container_endpoint_id"].(types.String)
	if originContainerEndpointId.IsUnknown() {
		return
	}

	originDnsPort := originAttr["dns_port"].(types.Int64)
	if originDnsPort.IsUnknown() {
		return
	}

	originDnsScheme := originAttr["dns_scheme"].(types.String)
	if originDnsScheme.IsUnknown() {
		return
	}

	hasMiddlewareScript := originMiddlewareScript.ValueInt64() > 0
	hasScriptingRoutingFilter := false
	isScriptingRoutingFiltersKnown := true

	if !routing.IsNull() {
		if routing.IsUnknown() {
			isScriptingRoutingFiltersKnown = false
		} else {
			routingFilters := routing.(types.Object).Attributes()["filters"]
			if routingFilters.IsUnknown() {
				isScriptingRoutingFiltersKnown = false
			} else {
				filterElements := routingFilters.(types.Set).Elements()

				var filters []string
				for _, filter := range filterElements {
					filters = append(filters, filter.(types.String).ValueString())
				}

				if slices.Contains(filters, "scripting") {
					hasScriptingRoutingFilter = true
				}
			}
		}
	}

	{
		if originType.ValueString() != "OriginUrl" {
			if !originUrl.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("url"), "Invalid origin.url value", "origin.url is only applicable for OriginUrl origins.")
			}

			if !originHostHeader.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("host_header"), "Invalid origin.host_header value", "origin.host_header is only applicable for OriginUrl origins.")
			}
		}

		if originType.ValueString() != "DnsAccelerate" {
			if !originDnsPort.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("dns_port"), "Invalid origin.dns_port value", "origin.dns_port is only applicable for DnsAccelerate origins.")
			}

			if !originDnsScheme.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("dns_scheme"), "Invalid origin.dns_scheme value", "origin.dns_scheme is only applicable for DnsAccelerate origins.")
			}
		}

		if originType.ValueString() != "StorageZone" {
			if !originStorageZone.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("storagezone"), "Invalid origin.storagezone value", "origin.storagezone is only applicable for StorageZone origins.")
			}
		}

		if originType.ValueString() != "ComputeScript" {
			if !originScript.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("script"), "Invalid origin.script value", "origin.script is only applicable for ComputeScript origins.")
			}
		}

		if originType.ValueString() != "ComputeContainer" {
			if !originContainerAppId.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("container_app_id"), "Invalid origin.container_app_id value", "origin.container_app_id is only applicable for ComputeContainer origins.")
			}

			if !originContainerEndpointId.IsNull() {
				resp.Diagnostics.AddAttributeError(originPath.AtName("container_endpoint_id"), "Invalid origin.container_endpoint_id value", "origin.container_endpoint_id is only applicable for ComputeContainer origins.")
			}
		}
	}

	if hasMiddlewareScript && !hasScriptingRoutingFilter {
		resp.Diagnostics.AddAttributeError(path.Root("routing").AtName("filters"), "Invalid routing.filters value", "middleware_script requires routing.filters to contain the \"scripting\" element.")
	}

	if hasScriptingRoutingFilter {
		if originType.ValueString() != "ComputeScript" && !hasMiddlewareScript {
			resp.Diagnostics.AddAttributeError(path.Root("routing").AtName("filters"), "Invalid routing.filters value", "The \"scripting\" filter must not be defined when scripts are not in use.")
		}
	}

	switch originType.ValueString() {
	case "OriginUrl":
		if originUrl.IsNull() || (!originUrl.IsUnknown() && originUrl.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid origin.url value", "OriginUrl requires origin.url to be defined.")
		} else if originUrl.ValueString() != "" {
			u, err := url.Parse(originUrl.ValueString())
			if err != nil {
				resp.Diagnostics.AddAttributeError(originTypePath.AtName("url"), "Invalid origin.url value", "origin.url must be a valid HTTP(s) URL.")
			} else if u.Scheme != "http" && u.Scheme != "https" {
				resp.Diagnostics.AddAttributeError(originTypePath.AtName("url"), "Invalid origin.url value", "origin.url must be a valid HTTP(s) URL.")
			}
		}

	case "DnsAccelerate":
		if originDnsPort.IsNull() || (!originDnsPort.IsUnknown() && originDnsPort.ValueInt64() <= 0) {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid origin.dns_port value", "DnsAccelerate requires origin.dns_port to be defined.")
		}

		if originDnsScheme.IsNull() || (!originDnsScheme.IsUnknown() && originDnsScheme.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid origin.dns_scheme value", "DnsAccelerate requires origin.dns_scheme to be defined.")
		}

	case "StorageZone":
		if originStorageZone.IsNull() || (!originStorageZone.IsUnknown() && originStorageZone.ValueInt64() <= 0) {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid origin.storagezone value", "StorageZone requires origin.storagezone to be defined.")
		}

	case "ComputeScript":
		if originScript.IsNull() || (!originScript.IsUnknown() && originScript.ValueInt64() <= 0) {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid origin.script value", "ComputeScript requires origin.script to be defined.")
		}

		if isScriptingRoutingFiltersKnown && !hasScriptingRoutingFilter {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid routing.filters value", "ComputeScript requires routing.filters to contain the \"scripting\" filter.")
		}

	case "ComputeContainer":
		if originContainerAppId.IsNull() || (!originContainerAppId.IsUnknown() && originContainerAppId.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid origin.container_app_id value", "ComputeContainer requires origin.container_app_id to be defined.")
		}

		if originContainerEndpointId.IsNull() || (!originContainerEndpointId.IsUnknown() && originContainerEndpointId.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(originTypePath, "Invalid origin.container_endpoint_id value", "ComputeContainer requires origin.container_endpoint_id to be defined.")
		}
	}
}
