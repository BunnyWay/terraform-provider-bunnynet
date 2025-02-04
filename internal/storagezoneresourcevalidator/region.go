package storagezoneresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
	"strings"
)

var standardRegions = []string{"BR", "DE", "JH", "LA", "NY", "SE", "SG", "SYD", "UK"}
var edgeRegions = []string{"BR", "CZ", "ES", "HK", "JH", "JP", "LA", "MI", "NY", "SE", "SG", "SYD", "UK", "WA"}

func Region() resource.ConfigValidator {
	return regionValidator{}
}

type regionValidator struct{}

func (v regionValidator) Description(ctx context.Context) string {
	return "Validate regions according to tier"
}

func (v regionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v regionValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	zoneTierAttr := path.Root("zone_tier")
	var zoneTier types.String
	request.Config.GetAttribute(ctx, zoneTierAttr, &zoneTier)

	regionAttr := path.Root("region")
	var region types.String
	request.Config.GetAttribute(ctx, regionAttr, &region)

	replicationRegionsAttr := path.Root("replication_regions")
	var replicationRegions types.Set
	request.Config.GetAttribute(ctx, replicationRegionsAttr, &replicationRegions)

	replicationRegionsElements := []string{}
	if !replicationRegions.IsNull() {
		for _, item := range replicationRegions.Elements() {
			replicationRegionsElements = append(replicationRegionsElements, item.(types.String).ValueString())
		}
	}

	if zoneTier.ValueString() == "Edge" {
		if region.ValueString() != "DE" {
			response.Diagnostics.AddAttributeError(regionAttr, "Invalid region attribute", "Storage zones in the Edge tier must have \"DE\" as the main region.")
			return
		}

		for _, item := range replicationRegionsElements {
			if slices.Contains(edgeRegions, item) {
				continue
			}

			response.Diagnostics.AddAttributeError(replicationRegionsAttr, "Invalid replication_regions attribute", fmt.Sprintf("\"%s\" is an invalid value. Must be one of: %s", item, strings.Join(edgeRegions, ", ")))
		}

		if response.Diagnostics.HasError() {
			return
		}
	}

	if zoneTier.ValueString() == "Standard" {
		// check if region matches standardRegions
		if !slices.Contains(standardRegions, region.ValueString()) {
			response.Diagnostics.AddAttributeError(regionAttr, "Invalid region attribute", fmt.Sprintf("\"%s\" is an invalid value. Must be one of: %s", region.ValueString(), strings.Join(standardRegions, ", ")))
			return
		}

		// check if replicationRegionsElements matches standardRegions
		for _, item := range replicationRegionsElements {
			if slices.Contains(standardRegions, item) {
				continue
			}

			response.Diagnostics.AddAttributeError(replicationRegionsAttr, "Invalid replication_regions attribute", fmt.Sprintf("\"%s\" is an invalid value. Must be one of: %s", item, strings.Join(standardRegions, ", ")))
		}

		if response.Diagnostics.HasError() {
			return
		}

		// check if region does not appear on replicationRegionsElements
		if slices.Contains(replicationRegionsElements, region.ValueString()) {
			response.Diagnostics.AddAttributeError(replicationRegionsAttr, "Invalid replication_regions attribute", fmt.Sprintf("\"%s\" is already defined as the \"region\" attribute", region.ValueString()))
			return
		}
	}
}
