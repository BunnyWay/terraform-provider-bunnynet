package computecontainerappresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EndpointNameShouldBeUnique() resource.ConfigValidator {
	return containerEndpointNameValidator{}
}

type containerEndpointNameValidator struct{}

func (v containerEndpointNameValidator) Description(ctx context.Context) string {
	return "Endpoint name must be unique"
}

func (v containerEndpointNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v containerEndpointNameValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	type endpointNameType struct {
		cIndex int
		eIndex int
	}

	containerAttr := path.Root("container")
	var containerSet types.Set
	request.Config.GetAttribute(ctx, containerAttr, &containerSet)
	endpointNames := map[string][]endpointNameType{}

	if len(containerSet.Elements()) == 0 {
		response.Diagnostics.AddError("No containers found", "No containers found")
		return
	}

	for cIndex, c := range containerSet.Elements() {
		for eIndex, e := range c.(types.Object).Attributes()["endpoint"].(types.Set).Elements() {
			name := e.(types.Object).Attributes()["name"].(types.String).ValueString()
			if _, ok := endpointNames[name]; !ok {
				endpointNames[name] = make([]endpointNameType, 0, 5)
			}
			endpointNames[name] = append(endpointNames[name], endpointNameType{cIndex, eIndex})
		}
	}

	for name, info := range endpointNames {
		if len(info) == 1 {
			continue
		}

		response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("container.endpoint.name should be unique. Found %d endpoints with name %s", len(info), name))
		return
	}
}
