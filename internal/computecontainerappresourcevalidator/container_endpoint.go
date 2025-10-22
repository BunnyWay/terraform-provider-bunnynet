package computecontainerappresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ContainerEndpoint() validator.Object {
	return containerEndpoint{}
}

const errEndpointCdnSslMissing = "There should be one \"origin_ssl\" configuration in the \"cdn\" block"
const errEndpointCdnNotNeeded = "There should be no \"cdn\" configuration for this endpoint type"
const errEndpointCdnMissing = "There should be one \"cdn\" configuration for this endpoint type"
const errEndpointCdnTooMany = "There should be only one \"cdn\" configuration for this endpoint type"
const errEndpointPortMissing = "There should be at least one \"port\" configuration for this endpoint type"
const errEndpointPortTooMany = "There should be only one \"port\" configuration for this endpoint type"
const errEndpointPortContainerMissing = "There should be at least one \"container\" port configured for this endpoint type"
const errEndpointPortExposedMissing = "There should be at least one \"exposed\" port configured for this endpoint type"
const errEndpointPortExposedNotNeeded = "There should be no \"exposed\" port configured for this endpoint type"
const errEndpointPortProtocolNotNeeded = "There should be no \"protocols\" configured for this endpoint type"
const errEndpointPortProtocolRequired = "There should be at least one \"protocols\" configured for this endpoint type"

type containerEndpoint struct{}

func (c containerEndpoint) Description(ctx context.Context) string {
	return "Validates endpoint configuration"
}

func (c containerEndpoint) MarkdownDescription(ctx context.Context) string {
	return c.Description(ctx)
}

func (c containerEndpoint) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	if request.ConfigValue.IsNull() {
		return
	}

	attrs := request.ConfigValue.Attributes()
	endpointType := attrs["type"].(types.String).ValueString()
	ports := attrs["port"].(types.List).Elements()
	cdn := attrs["cdn"].(types.List).Elements()
	errorSummary := fmt.Sprintf("Invalid endpoint.%s configuration", attrs["name"].(types.String).ValueString())

	if endpointType == "CDN" && len(cdn) == 0 {
		response.Diagnostics.AddError(errorSummary, errEndpointCdnMissing)
		return
	}

	if endpointType == "CDN" && len(cdn) > 1 {
		response.Diagnostics.AddError(errorSummary, errEndpointCdnTooMany)
		return
	}

	if endpointType != "CDN" && len(cdn) > 0 {
		response.Diagnostics.AddError(errorSummary, errEndpointCdnNotNeeded)
		return
	}

	if endpointType == "CDN" && len(cdn) == 1 {
		cdnAttrs := cdn[0].(types.Object).Attributes()

		if v, ok := cdnAttrs["origin_ssl"]; !ok || v.IsNull() {
			response.Diagnostics.AddError(errorSummary, errEndpointCdnSslMissing)
			return
		}
	}

	if len(ports) == 0 {
		response.Diagnostics.AddError(errorSummary, errEndpointPortMissing)
		return
	}

	if endpointType != "Anycast" && len(ports) > 1 {
		response.Diagnostics.AddError(errorSummary, errEndpointPortTooMany)
		return
	}

	for _, port := range ports {
		portAttr := port.(types.Object).Attributes()

		// container port
		if portAttr["container"].IsNull() {
			response.Diagnostics.AddError(errorSummary, errEndpointPortContainerMissing)
			return
		}

		// exposed port
		if endpointType == "Anycast" && portAttr["exposed"].IsNull() {
			response.Diagnostics.AddError(errorSummary, errEndpointPortExposedMissing)
			return
		}

		if endpointType != "Anycast" && !portAttr["exposed"].IsNull() {
			response.Diagnostics.AddError(errorSummary, errEndpointPortExposedNotNeeded)
			return
		}

		// protocols
		var protocols []string
		diags := portAttr["protocols"].(types.Set).ElementsAs(ctx, &protocols, false)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		if (endpointType == "Anycast" || endpointType == "InternalIP") && len(protocols) == 0 {
			response.Diagnostics.AddError(errorSummary, errEndpointPortProtocolRequired)
			return
		}

		if endpointType == "CDN" && len(protocols) > 0 {
			response.Diagnostics.AddError(errorSummary, errEndpointPortProtocolNotNeeded)
			return
		}
	}
}
