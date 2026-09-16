// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

// This file was generated via "go generate". DO NOT EDIT.
package pullzoneshieldresourcevalidator

const BotCategorizationBotActionOptBlock = 1
const BotCategorizationBotActionOptAllow = 2
const BotCategorizationBotActionOptIgnore = 3

var BotCategorizationBotActionMap = map[uint8]string{
	BotCategorizationBotActionOptBlock:  "Block",
	BotCategorizationBotActionOptAllow:  "Allow",
	BotCategorizationBotActionOptIgnore: "Ignore",
}

var BotCategorizationCategoryMap = map[uint8]string{
	1: "SEO",
	2: "AIScraper",
	3: "AITool",
	4: "Tool",
	5: "Ads",
	6: "Preview",
	7: "Social",
}

var PlanTypeMap = map[uint8]string{
	0: "Basic",
	1: "Advanced",
	2: "Business",
	3: "Enterprise",
}
