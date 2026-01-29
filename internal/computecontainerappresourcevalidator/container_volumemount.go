package computecontainerappresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ContainerVolumeMounts() resource.ConfigValidator {
	return containerVolumeMounts{}
}

type containerVolumeMounts struct {
}

func (v containerVolumeMounts) Description(ctx context.Context) string {
	return "Volumes can only be mounted once"
}

func (v containerVolumeMounts) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v containerVolumeMounts) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	containerAttr := path.Root("container")
	var containerList types.List
	request.Config.GetAttribute(ctx, containerAttr, &containerList)
	volumeMounts := map[string]uint8{}
	pathMounts := map[string]uint8{}

	if len(containerList.Elements()) == 0 {
		response.Diagnostics.AddError("No containers found", "No containers found")
		return
	}

	for _, c := range containerList.Elements() {
		for _, e := range c.(types.Object).Attributes()["volumemount"].(types.List).Elements() {
			mountName := e.(types.Object).Attributes()["name"].(types.String).ValueString()
			if _, ok := volumeMounts[mountName]; !ok {
				volumeMounts[mountName] = 0
			}
			volumeMounts[mountName]++

			mountPath := e.(types.Object).Attributes()["path"].(types.String).ValueString()
			if _, ok := pathMounts[mountPath]; !ok {
				pathMounts[mountPath] = 0
			}
			pathMounts[mountPath]++
		}
	}

	volumeAttr := path.Root("volume")
	var volumeList types.List
	request.Config.GetAttribute(ctx, volumeAttr, &volumeList)
	volumes := map[string]struct{}{}

	for _, v := range volumeList.Elements() {
		name := v.(types.Object).Attributes()["name"].(types.String).ValueString()
		volumes[name] = struct{}{}
	}

	for name, count := range volumeMounts {
		if _, ok := volumes[name]; !ok {
			response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("The volume \"%s\" does not exist", name))
			continue
		}

		if count > 1 {
			response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("The volume \"%s\" can only be mounted once", name))
			continue
		}
	}

	for pathMount, count := range pathMounts {
		if count > 1 {
			response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("There are multiple volumes mounted at the path \"%s\"", pathMount))
			continue
		}
	}
}
