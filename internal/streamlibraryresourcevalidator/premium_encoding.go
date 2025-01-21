package streamlibraryresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PremiumEncoding() resource.ConfigValidator {
	return premiumEncodingValidator{}
}

type premiumEncodingValidator struct{}

func (v premiumEncodingValidator) Description(ctx context.Context) string {
	return "Premium Encoding features"
}

func (v premiumEncodingValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v premiumEncodingValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	encodingTierAttr := path.Root("encoding_tier")
	var encodingTier types.String
	request.Config.GetAttribute(ctx, encodingTierAttr, &encodingTier)
	isPremium := !encodingTier.IsNull() && encodingTier.ValueString() == "Premium"

	// jit_encoding
	{
		jitAttr := path.Root("jit_encoding")
		var jitEncoding types.Bool
		request.Config.GetAttribute(ctx, jitAttr, &jitEncoding)

		earlyPlayAttr := path.Root("early_play_enabled")
		var earlyPlay types.Bool
		request.Config.GetAttribute(ctx, earlyPlayAttr, &earlyPlay)

		if jitEncoding.ValueBool() && !isPremium {
			response.Diagnostics.AddAttributeError(jitAttr, "Premium encoding is not enabled", fmt.Sprintf("\"%s\" is part of Premium Encoding. You must configure the \"%s\" attribute.", jitAttr, encodingTierAttr))
		}

		if jitEncoding.ValueBool() && earlyPlay.ValueBool() {
			response.Diagnostics.AddAttributeError(jitAttr, "Incompatible Attribute Combination", fmt.Sprintf("\"%s\" and \"%s\" cannot be enabled simultaneously.", jitAttr, earlyPlayAttr))
			response.Diagnostics.AddAttributeError(earlyPlayAttr, "Incompatible Attribute Combination", fmt.Sprintf("\"%s\" and \"%s\" cannot be enabled simultaneously.", jitAttr, earlyPlayAttr))
		}
	}

	// output_codecs
	{
		outputCodecsAttr := path.Root("output_codecs")
		var outputCodecs types.Set
		request.Config.GetAttribute(ctx, outputCodecsAttr, &outputCodecs)

		for _, codec := range outputCodecs.Elements() {
			value := codec.(types.String).ValueString()

			if !isPremium && value == "vp9" {
				response.Diagnostics.AddAttributeError(outputCodecsAttr, "Premium encoding is not enabled", fmt.Sprintf("\"%s\" is part of Premium Encoding. You must configure the \"%s\" attribute.", value, encodingTierAttr))
			}
		}
	}
}
