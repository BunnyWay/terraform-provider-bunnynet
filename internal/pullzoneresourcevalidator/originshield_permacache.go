// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package pullzoneresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func OriginShieldPermacache() resource.ConfigValidator {
	return originShieldPermacacheValidator{}
}

type originShieldPermacacheValidator struct{}

func (v originShieldPermacacheValidator) Description(ctx context.Context) string {
	return "Origin Shield and Perma-Cache are mutually exclusive."
}

func (v originShieldPermacacheValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v originShieldPermacacheValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	originShieldEnabledAttr := path.Root("originshield_enabled")
	permacacheStoragezoneAttr := path.Root("permacache_storagezone")

	var originShieldEnabled types.Bool
	request.Config.GetAttribute(ctx, originShieldEnabledAttr, &originShieldEnabled)

	if originShieldEnabled.IsUnknown() || originShieldEnabled.IsNull() || !originShieldEnabled.ValueBool() {
		return
	}

	var permacacheStoragezone types.Int64
	request.Config.GetAttribute(ctx, permacacheStoragezoneAttr, &permacacheStoragezone)

	if permacacheStoragezone.IsNull() {
		return
	}

	if !permacacheStoragezone.IsUnknown() && permacacheStoragezone.ValueInt64() == 0 {
		return
	}

	response.Diagnostics.AddAttributeError(originShieldEnabledAttr, "Conflicting attributes", fmt.Sprintf(`Origin Shield and Perma-Cache are mutually exclusive. When "%s" is set, "%s" must be false or omitted.`, permacacheStoragezoneAttr.String(), originShieldEnabledAttr.String()))
}
