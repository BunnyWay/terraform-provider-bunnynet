package pullzonehostnameresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CustomCertificate() resource.ConfigValidator {
	return customCertificateValidator{}
}

type customCertificateValidator struct{}

func (v customCertificateValidator) Description(ctx context.Context) string {
	return "When using custom certificates, the tls_enabled must be true."
}

func (v customCertificateValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v customCertificateValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var certificate types.String
	request.Config.GetAttribute(ctx, path.Root("certificate"), &certificate)

	if certificate.IsNull() || len(certificate.ValueString()) == 0 {
		return
	}

	tlsEnabledAttr := path.Root("tls_enabled")
	var tlsEnabled types.Bool
	request.Config.GetAttribute(ctx, tlsEnabledAttr, &tlsEnabled)

	if tlsEnabled.ValueBool() {
		return
	}

	response.Diagnostics.AddAttributeError(tlsEnabledAttr, "Attribute must be set to true", fmt.Sprintf("\"%s\" must be set to true when defining custom certificates.", tlsEnabledAttr))
}
