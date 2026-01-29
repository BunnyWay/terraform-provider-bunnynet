package computecontainerappresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func VolumeNamesShouldBeUnique() resource.ConfigValidator {
	return volumeNameValidator{}
}

type volumeNameValidator struct{}

func (v volumeNameValidator) Description(ctx context.Context) string {
	return "Volume names should be unique"
}

func (v volumeNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v volumeNameValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	volumeAttr := path.Root("volume")
	var volumeList types.List
	request.Config.GetAttribute(ctx, volumeAttr, &volumeList)
	volumes := map[string]struct{}{}

	for _, v := range volumeList.Elements() {
		name := v.(types.Object).Attributes()["name"].(types.String).ValueString()
		if _, ok := volumes[name]; ok {
			response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("A volume with name \"%s\" was already declared", name))
			continue
		}

		volumes[name] = struct{}{}
	}
}
