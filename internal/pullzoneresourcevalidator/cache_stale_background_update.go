package pullzoneresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CacheStaleBackgroundUpdate() resource.ConfigValidator {
	return cacheStaleBackgroundUpdateValidator{}
}

type cacheStaleBackgroundUpdateValidator struct{}

func (v cacheStaleBackgroundUpdateValidator) Description(ctx context.Context) string {
	return "When using cache_stale, the use_background_update must be true."
}

func (v cacheStaleBackgroundUpdateValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v cacheStaleBackgroundUpdateValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var cacheStale types.Set
	request.Config.GetAttribute(ctx, path.Root("cache_stale"), &cacheStale)

	if cacheStale.IsUnknown() || cacheStale.IsNull() || len(cacheStale.Elements()) == 0 {
		return
	}

	attr := path.Root("use_background_update")
	var useBackgroundUpdate types.Bool
	request.Config.GetAttribute(ctx, attr, &useBackgroundUpdate)

	if useBackgroundUpdate.IsUnknown() || useBackgroundUpdate.IsNull() || useBackgroundUpdate.ValueBool() {
		return
	}

	response.Diagnostics.AddAttributeError(attr, "Attribute must be default", fmt.Sprintf("\"%s\" must be true. It can also be omitted.", attr.String()))
}
