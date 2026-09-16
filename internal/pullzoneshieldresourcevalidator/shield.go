// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneshieldresourcevalidator

import "github.com/bunnyway/terraform-provider-bunnynet/internal/utils"

var BotCategorizationBotActionMapInverted = utils.MapInvert(BotCategorizationBotActionMap)
var BotCategorizationCategoryActionMapInverted = utils.MapInvert(BotCategorizationCategoryActionMap)
var BotCategorizationCategoryMapInverted = utils.MapInvert(BotCategorizationCategoryMap)

const BotCategorizationCategoryActionOptIgnore = 0
const BotCategorizationCategoryActionOptBlock = 1
const BotCategorizationCategoryActionOptAllow = 2

var BotCategorizationCategoryActionMap = map[uint8]string{
	BotCategorizationCategoryActionOptIgnore: "Ignore",
	BotCategorizationCategoryActionOptBlock:  "Block",
	BotCategorizationCategoryActionOptAllow:  "Allow",
}
