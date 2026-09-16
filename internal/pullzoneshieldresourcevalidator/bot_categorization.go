// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneshieldresourcevalidator

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
	"slices"
	"strings"
)

var _ validator.List = botCategorization{}

func BotCategorization() validator.List {
	return botCategorization{}
}

type botCategorization struct {
}

func (v botCategorization) Description(ctx context.Context) string {
	return "Validate bot_categorization list"
}

func (v botCategorization) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v botCategorization) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	items := request.ConfigValue.Elements()
	categories := make([]string, 0, len(BotCategorizationCategoryMap))

	for n, item := range items {
		catAttrs := item.(types.Object).Attributes()
		catName := catAttrs["category"].(types.String).ValueString()
		overrideEl := catAttrs["overrides"].(types.List)

		categories = append(categories, catName)

		if overrideEl.IsNull() || overrideEl.IsUnknown() {
			continue
		}

		catAction := catAttrs["action"].(types.String).ValueString()
		botNames := make([]string, len(overrideEl.Elements()))

		for i, bot := range overrideEl.Elements() {
			botAttrs := bot.(types.Object).Attributes()
			botName := botAttrs["bot"].(types.String).ValueString()
			botAction := botAttrs["action"].(types.String).ValueString()

			if botName == "" {
				response.Diagnostics.AddAttributeError(request.Path.AtListIndex(n), "Invalid override bot", "Unexpected empty name for override bot")
				return
			}

			if botAction == catAction {
				response.Diagnostics.AddAttributeError(request.Path.AtListIndex(n), "Invalid override action", fmt.Sprintf(`The override action for "%s" is the same as the category action. Please remove the override.`, botName))
				return
			}

			if _, ok := BotCategorizationBotActionMapInverted[botAction]; !ok {
				p := path.Root("overrides").AtListIndex(i)
				response.Diagnostics.AddAttributeError(request.Path.AtListIndex(n), "Invalid override action", fmt.Sprintf(`Attribute %s value must be one of: ["%s"], got: "%s"`, p, strings.Join(maps.Values(BotCategorizationBotActionMap), `", "`), botAction))
				return
			}

			botNames[i] = botName
		}

		if !slices.IsSorted(botNames) {
			response.Diagnostics.AddAttributeError(request.Path.AtListIndex(n), "Invalid override list", "To avoid errors after applying a plan, please sort the list alphabetically by the `bot` attribute.")
			return
		}
	}

	if !slices.IsSorted(categories) {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid bot_categorization list", "To avoid errors after applying a plan, please sort the list alphabetically by the `category` attribute.")
		return
	}

	validCategories := maps.Values(BotCategorizationCategoryMap)
	diff := utils.SliceDiff(validCategories, categories)

	if len(diff) > 0 {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid bot_categorization list", fmt.Sprintf("The list is missing the following categories: %+v", diff))
		return
	}

	extra := utils.SliceDiff(categories, validCategories)
	if len(extra) > 0 {
		response.Diagnostics.AddAttributeError(request.Path, "Invalid bot_categorization list", fmt.Sprintf("The list has unexpected categories: %+v", extra))
		return
	}
}
