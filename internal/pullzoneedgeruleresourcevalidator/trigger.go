package pullzoneedgeruleresourcevalidator

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func TriggerObject() resource.ConfigValidator {
	return triggerObjectValidator{}
}

type triggerObjectValidator struct{}

func (v triggerObjectValidator) Description(ctx context.Context) string {
	return "The trigger object must be valid."
}

func (v triggerObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v triggerObjectValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	triggerPath := path.Root("triggers")

	var triggers types.List
	request.Config.GetAttribute(ctx, triggerPath, &triggers)

	for _, trigger := range triggers.Elements() {
		triggerAttrs := trigger.(types.Object).Attributes()

		// trigger.match_type
		{
			matchType := triggerAttrs["match_type"].(types.String).ValueString()
			values := maps.Values(TriggerMatchTypeMap)

			if !slices.Contains(values, matchType) {
				response.Diagnostics.AddAttributeError(
					triggerPath,
					"Trigger match_type must be valid",
					fmt.Sprintf("Trigger \"match_type\" must be one of %s.", utils.StringSliceToTerraformSet(values)),
				)
				return
			}
		}

		// trigger.patterns
		{
			patterns := triggerAttrs["patterns"].(types.List).Elements()

			if len(patterns) < 1 {
				response.Diagnostics.AddAttributeError(
					triggerPath,
					"Trigger patterns must be valid",
					"Trigger \"patterns\" must have at least one element.",
				)
				return
			}
		}

		// trigger.type
		{
			triggerType := triggerAttrs["type"].(types.String).ValueString()
			values := maps.Values(TriggerTypeMap)

			if !slices.Contains(values, triggerType) {
				response.Diagnostics.AddAttributeError(
					triggerPath,
					"Trigger type must be valid",
					fmt.Sprintf("Trigger \"type\" must be one of %s.", utils.StringSliceToTerraformSet(values)),
				)
				return
			}
		}
	}
}
