package boolvalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func RequiresAttributeAsTrue(attr string) validator.Bool {
	return requiresAttributeAsTrue{attr}
}

type requiresAttributeAsTrue struct {
	attributeName string
}

func (v requiresAttributeAsTrue) Description(ctx context.Context) string {
	return fmt.Sprintf("Requires attribute %s to be true.", v.attributeName)
}

func (v requiresAttributeAsTrue) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v requiresAttributeAsTrue) ValidateBool(ctx context.Context, request validator.BoolRequest, response *validator.BoolResponse) {
	attr1Value := request.ConfigValue.ValueBool()

	var attr2Value bool
	request.Config.GetAttribute(ctx, path.Root(v.attributeName), &attr2Value)

	if attr1Value && !attr2Value {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Incompatible Attribute Combination",
			fmt.Sprintf("Attribute \"%s\" must also be set to true.", v.attributeName),
		))
	}
}
