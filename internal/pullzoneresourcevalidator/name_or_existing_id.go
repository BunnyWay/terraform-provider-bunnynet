// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NameOrExistingId() resource.ConfigValidator {
	return nameOrExistingIdValidator{}
}

type nameOrExistingIdValidator struct{}

func (v nameOrExistingIdValidator) Description(ctx context.Context) string {
	return "Either \"name\" or \"existing_id\" must be set."
}

func (v nameOrExistingIdValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v nameOrExistingIdValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var name types.String
	req.Config.GetAttribute(ctx, path.Root("name"), &name)

	var existingId types.Int64
	req.Config.GetAttribute(ctx, path.Root("existing_id"), &existingId)

	if name.IsUnknown() || existingId.IsUnknown() {
		return
	}

	nameSet := !name.IsNull()
	existingIdSet := !existingId.IsNull()

	if !nameSet && !existingIdSet {
		resp.Diagnostics.AddError(
			"Missing required attribute",
			"Either \"name\" or \"existing_id\" must be configured.",
		)
	}
}
