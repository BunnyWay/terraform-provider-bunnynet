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
	var permacacheStoragezone types.Int64
	request.Config.GetAttribute(ctx, path.Root("permacache_storagezone"), &permacacheStoragezone)

	if permacacheStoragezone.IsNull() {
		return
	}

	attr := path.Root("cache_expiration_time")
	var cacheExpirationTime types.Int64
	request.Config.GetAttribute(ctx, attr, &cacheExpirationTime)

	if cacheExpirationTime.IsNull() || cacheExpirationTime.ValueInt64() == DefaultCacheExpirationTimeForPermacache {
		return
	}

	response.Diagnostics.AddAttributeError(attr, "Attribute must be default", fmt.Sprintf("\"%s\" must be %d. It can also be omitted.", attr.String(), DefaultCacheExpirationTimeForPermacache))
}
