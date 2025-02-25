package computecontainerappresourcevalidator

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func RegionRequiredMustAlsoBeAllowed() resource.ConfigValidator {
	return regionRequiredMustAlsoBeAllowedValidator{}
}

type regionRequiredMustAlsoBeAllowedValidator struct{}

func (v regionRequiredMustAlsoBeAllowedValidator) Description(ctx context.Context) string {
	return "Any region required region must also be defined as an allowed region."
}

func (v regionRequiredMustAlsoBeAllowedValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v regionRequiredMustAlsoBeAllowedValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	allowedAttr := path.Root("regions_allowed")
	requiredAttr := path.Root("regions_required")
	var allowedRegions, requiredRegions []string

	{
		var attrSet types.Set
		request.Config.GetAttribute(ctx, allowedAttr, &attrSet)
		allowedRegions = utils.ConvertSetToStringSlice(attrSet)
	}

	{
		var attrSet types.Set
		request.Config.GetAttribute(ctx, requiredAttr, &attrSet)
		requiredRegions = utils.ConvertSetToStringSlice(attrSet)
	}

	diff := utils.SliceDiff(requiredRegions, allowedRegions)
	if len(diff) > 0 {
		missingValues := make([]attr.Value, len(diff))
		for i, el := range diff {
			missingValues[i] = types.StringValue(el)
		}

		missingValuesSet, diags := types.SetValue(types.StringType, missingValues)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		response.Diagnostics.AddAttributeError(allowedAttr, "Incompatible Attribute Combination", fmt.Sprintf("\"%s\" must also contain all required regions. Missing: %s", allowedAttr, missingValuesSet))
		return
	}
}
