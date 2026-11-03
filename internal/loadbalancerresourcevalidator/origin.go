// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package loadbalancerresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/customtype"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/url"
)

func Origin() resource.ConfigValidator {
	return originValidator{}
}

type originValidator struct{}

func (v originValidator) Description(ctx context.Context) string {
	return "Validations for Load Balancer origin type"
}

func (v originValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v originValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	typePath := path.Root("type")

	var originType types.String
	req.Config.GetAttribute(ctx, typePath, &originType)
	if originType.IsUnknown() {
		return
	}

	var originUrl customtype.PullzoneOriginUrlValue
	req.Config.GetAttribute(ctx, path.Root("url"), &originUrl)
	if originUrl.IsUnknown() {
		return
	}

	var originHostHeader types.String
	req.Config.GetAttribute(ctx, path.Root("host_header"), &originHostHeader)
	if originHostHeader.IsUnknown() {
		return
	}

	var originStorageZone types.Int64
	req.Config.GetAttribute(ctx, path.Root("storagezone"), &originStorageZone)
	if originStorageZone.IsUnknown() {
		return
	}

	var originScript types.Int64
	req.Config.GetAttribute(ctx, path.Root("script"), &originScript)
	if originScript.IsUnknown() {
		return
	}

	var originContainerAppId types.String
	req.Config.GetAttribute(ctx, path.Root("container_app_id"), &originContainerAppId)
	if originContainerAppId.IsUnknown() {
		return
	}

	var originContainerEndpointId types.String
	req.Config.GetAttribute(ctx, path.Root("container_endpoint_id"), &originContainerEndpointId)
	if originContainerEndpointId.IsUnknown() {
		return
	}

	{
		if originType.ValueString() != "OriginUrl" {
			if !originUrl.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("url"), "Invalid url value", `"url" is only applicable for OriginUrl origins.`)
			}

			if !originHostHeader.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("host_header"), "Invalid host_header value", `"host_header" is only applicable for OriginUrl origins.`)
			}
		}

		if originType.ValueString() != "StorageZone" {
			if !originStorageZone.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("storagezone"), "Invalid storagezone value", `"storagezone" is only applicable for StorageZone origins.`)
			}
		}

		if originType.ValueString() != "ComputeScript" {
			if !originScript.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("script"), "Invalid script value", `"script" is only applicable for ComputeScript origins.`)
			}
		}

		if originType.ValueString() != "ComputeContainer" {
			if !originContainerAppId.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("container_app_id"), "Invalid container_app_id value", `"container_app_id" is only applicable for ComputeContainer origins.`)
			}

			if !originContainerEndpointId.IsNull() {
				resp.Diagnostics.AddAttributeError(path.Root("container_endpoint_id"), "Invalid container_endpoint_id value", `"container_endpoint_id" is only applicable for ComputeContainer origins.`)
			}
		}
	}

	switch originType.ValueString() {
	case "OriginUrl":
		if originUrl.IsNull() || (!originUrl.IsUnknown() && originUrl.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(typePath, "Invalid url value", `OriginUrl requires attribute "url"" to be defined.`)
		} else if originUrl.ValueString() != "" {
			u, err := url.Parse(originUrl.ValueString())
			if err != nil {
				resp.Diagnostics.AddAttributeError(path.Root("url"), "Invalid url value", `Value must be a valid HTTP(s) URL.`)
			} else if u.Scheme != "http" && u.Scheme != "https" {
				resp.Diagnostics.AddAttributeError(path.Root("url"), "Invalid url value", `Value must be a valid HTTP(s) URL.`)
			}
		}

	case "StorageZone":
		if originStorageZone.IsNull() || (!originStorageZone.IsUnknown() && originStorageZone.ValueInt64() <= 0) {
			resp.Diagnostics.AddAttributeError(typePath, "Invalid storagezone value", `StorageZone requires attribute "storagezone" to be defined.`)
		}

	case "ComputeScript":
		if originScript.IsNull() || (!originScript.IsUnknown() && originScript.ValueInt64() <= 0) {
			resp.Diagnostics.AddAttributeError(typePath, "Invalid script value", `ComputeScript requires attribute "script" to be defined.`)
		}

	case "ComputeContainer":
		if originContainerAppId.IsNull() || (!originContainerAppId.IsUnknown() && originContainerAppId.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(typePath, "Invalid container_app_id value", `ComputeContainer requires attribute "container_app_id" to be defined.`)
		}

		if originContainerEndpointId.IsNull() || (!originContainerEndpointId.IsUnknown() && originContainerEndpointId.ValueString() == "") {
			resp.Diagnostics.AddAttributeError(typePath, "Invalid container_endpoint_id value", `ComputeContainer requires attribute "container_endpoint_id" to be defined.`)
		}
	}
}
