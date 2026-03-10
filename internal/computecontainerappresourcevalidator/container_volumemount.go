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
	volumeMounts := map[string]map[string]uint8{}
	pathMounts := map[string]map[string]uint8{}

	if len(containerList.Elements()) == 0 {
		response.Diagnostics.AddError("No containers found", "No containers found")
		return
	}

	for _, c := range containerList.Elements() {
		cAttrs := c.(types.Object).Attributes()
		cName := cAttrs["name"].(types.String).ValueString()

		for _, e := range cAttrs["volumemount"].(types.List).Elements() {
			// name
			{
				mountName := e.(types.Object).Attributes()["name"].(types.String).ValueString()

				if _, ok := volumeMounts[mountName]; !ok {
					volumeMounts[mountName] = map[string]uint8{}
				}

				if _, ok := volumeMounts[mountName][cName]; !ok {
					volumeMounts[mountName][cName] = 0
				}

				volumeMounts[mountName][cName]++
			}

			// path
			{
				mountPath := e.(types.Object).Attributes()["path"].(types.String).ValueString()

				if _, ok := pathMounts[cName]; !ok {
					pathMounts[cName] = map[string]uint8{}
				}

				if _, ok := pathMounts[cName][mountPath]; !ok {
					pathMounts[cName][mountPath] = 0
				}

				pathMounts[cName][mountPath]++
			}
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

	for name, containerMounts := range volumeMounts {
		if _, ok := volumes[name]; !ok {
			response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("The volume \"%s\" does not exist", name))
			continue
		}

		for containerName, count := range containerMounts {
			if count > 1 {
				response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("The volume \"%s\" can only be mounted once on container \"%s\"", name, containerName))
				continue
			}
		}
	}

	for containerName, containerMounts := range pathMounts {
		for pathMount, count := range containerMounts {
			if count > 1 {
				response.Diagnostics.AddError("Invalid endpoint configuration", fmt.Sprintf("There are multiple volumes mounted at the path \"%s\" on container \"%s\"", pathMount, containerName))
				continue
			}
		}
	}
}
