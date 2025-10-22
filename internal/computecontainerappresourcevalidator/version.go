package computecontainerappresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func Version() validator.Int64 {
	return version{}
}

type version struct{}

func (v version) Description(ctx context.Context) string {
	return "Make sure the resource addresses backwards compatibility breaks."
}

func (v version) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v version) ValidateInt64(ctx context.Context, request validator.Int64Request, response *validator.Int64Response) {
	if request.ConfigValue.IsUnknown() {
		return
	}

	if request.ConfigValue.IsNull() || request.ConfigValue.ValueInt64() != 2 {
		response.Diagnostics.AddError("Missing version attribute", "This resource had a backwards incompatible change in v0.11.0.\n\nPlease make sure to read through https://github.com/BunnyWay/terraform-provider-bunnynet/releases/tag/v0.11.0\n\nTo suppress this error message, add `version = 2` to this resource.")
	}
}
