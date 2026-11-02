// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package computecontainerimageresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

func Registry() validator.String {
	return registryValidator{}
}

type registryValidator struct{}

func (v registryValidator) Description(ctx context.Context) string {
	return "Registry type"
}

func (v registryValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v registryValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	tfRegistry := req.ConfigValue.ValueString()

	if tfRegistry == "GitHub" || tfRegistry == "DockerHub" {
		return
	}

	re := regexp.MustCompile(`^([a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+(:([0-9]+))?)$`)
	if !re.Match([]byte(tfRegistry)) {
		resp.Diagnostics.AddError("Invalid registry", "Registry must be a hostname, with an optional port.")
	}
}
