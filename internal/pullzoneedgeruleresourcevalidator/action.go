package pullzoneedgeruleresourcevalidator

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

func ActionParameters() resource.ConfigValidator {
	return actionParameters{}
}

type actionParameters struct{}

func (v actionParameters) Description(ctx context.Context) string {
	return "The action parameters must be valid for the action type"
}

func (v actionParameters) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v actionParameters) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	actionsPath := path.Root("actions")

	var actions types.List
	request.Config.GetAttribute(ctx, actionsPath, &actions)

	if !actions.IsNull() {
		for _, action := range actions.Elements() {
			actionAttr := action.(types.Object).Attributes()

			if actionAttr["parameter1"].IsUnknown() || actionAttr["parameter2"].IsUnknown() || actionAttr["parameter3"].IsUnknown() {
				return
			}

			err := v.validateAction(
				actionAttr["type"].(types.String).ValueString(),
				actionAttr["parameter1"].(types.String).ValueString(),
				actionAttr["parameter2"].(types.String).ValueString(),
				actionAttr["parameter3"].(types.String).ValueString(),
			)

			if err != nil {
				response.Diagnostics.AddAttributeError(actionsPath, "Invalid attribute configuration", err.Error())
				return
			}
		}
	} else {
		actionPath := path.Root("action")

		var actionType types.String
		var actionParam1 types.String
		var actionParam2 types.String
		var actionParam3 types.String

		request.Config.GetAttribute(ctx, actionPath, &actionType)
		request.Config.GetAttribute(ctx, path.Root("action_parameter1"), &actionParam1)
		request.Config.GetAttribute(ctx, path.Root("action_parameter2"), &actionParam2)
		request.Config.GetAttribute(ctx, path.Root("action_parameter3"), &actionParam3)

		if actionParam1.IsUnknown() || actionParam2.IsUnknown() || actionParam3.IsUnknown() {
			return
		}

		err := v.validateAction(actionType.ValueString(), actionParam1.ValueString(), actionParam2.ValueString(), actionParam3.ValueString())
		if err != nil {
			response.Diagnostics.AddAttributeError(actionPath, "Invalid attribute configuration", err.Error())
			return
		}
	}
}

func (v actionParameters) validateAction(action string, parameter1 string, parameter2 string, parameter3 string) error {
	switch action {
	case "Redirect":
		if len(parameter1) == 0 {
			return errors.New("parameter1 must not be empty")
		}

		if len(parameter2) == 0 {
			return errors.New("parameter2 must not be empty")
		}

		options := []string{"301", "302", "307", "308"}
		if !slices.Contains(options, parameter2) {
			return fmt.Errorf("parameter2 must be one of: %s", options)
		}
	}

	return nil
}
