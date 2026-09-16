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

const DefaultCacheExpirationTimeForPermacache = 31919000

func PermacacheCacheExpirationTime() resource.ConfigValidator {
	return permacacheCacheExpirationTimeValidator{}
}

type permacacheCacheExpirationTimeValidator struct{}

func (v permacacheCacheExpirationTimeValidator) Description(ctx context.Context) string {
	return "When using permacache, the cache_expiration_time must be set to 1y."
}

func (v permacacheCacheExpirationTimeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v permacacheCacheExpirationTimeValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	cacheExpirationTimeAttr := path.Root("cache_expiration_time")
	permacacheStoragezoneAttr := path.Root("permacache_storagezone")

	var cacheExpirationTime types.Int64
	request.Config.GetAttribute(ctx, cacheExpirationTimeAttr, &cacheExpirationTime)

	var permacacheStoragezone types.Int64
	request.Config.GetAttribute(ctx, permacacheStoragezoneAttr, &permacacheStoragezone)

	if permacacheStoragezone.IsNull() || cacheExpirationTime.IsUnknown() || cacheExpirationTime.IsNull() {
		return
	}

	if permacacheStoragezone.IsUnknown() && cacheExpirationTime.IsUnknown() {
		return
	}

	permacacheStoragezoneId := permacacheStoragezone.ValueInt64()

	if !permacacheStoragezone.IsUnknown() && !permacacheStoragezone.IsNull() && permacacheStoragezoneId == 0 {
		return
	}

	if permacacheStoragezoneId > 0 && cacheExpirationTime.ValueInt64() == DefaultCacheExpirationTimeForPermacache {
		return
	}

	response.Diagnostics.AddAttributeError(cacheExpirationTimeAttr, "Attribute must be omitted", fmt.Sprintf(`When "%s" is defined, "%s" should be omitted.`, permacacheStoragezoneAttr.String(), cacheExpirationTimeAttr.String()))
}
