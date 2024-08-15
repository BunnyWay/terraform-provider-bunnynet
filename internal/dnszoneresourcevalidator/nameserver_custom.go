package dnszoneresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CustomNameserver() resource.ConfigValidator {
	return customNameserverValidator{}
}

type customNameserverValidator struct{}

func (v customNameserverValidator) Description(ctx context.Context) string {
	return "When using custom nameservers, the nameservers must be defined."
}

func (v customNameserverValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v customNameserverValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var nameserverCustom types.Bool
	request.Config.GetAttribute(ctx, path.Root("nameserver_custom"), &nameserverCustom)
	mustBeDefault := !nameserverCustom.ValueBool()

	attrs := []path.Path{
		path.Root("nameserver1"),
		path.Root("nameserver2"),
		path.Root("soa_email"),
	}

	for _, attr := range attrs {
		attrDefault, err := v.getStringDefault(ctx, request, attr)
		if err != nil {
			response.Diagnostics.AddAttributeError(attr, "Could not obtain the default value", err.Error())
			return
		}

		attrType := v.getStringValue(ctx, request, attr)
		if attrType.IsNull() {
			attrDefault = attrType.ValueString()
		}

		if mustBeDefault && (attrType.ValueString() != attrDefault) {
			response.Diagnostics.AddAttributeError(attr, "Attribute must be default", "\""+attr.String()+"\" must be equal to the default value")
			continue
		}

		if !mustBeDefault && attrType.ValueString() == attrDefault {
			response.Diagnostics.AddAttributeError(attr, "Attribute must be defined", "\""+attr.String()+"\" must be different than the default value")
			continue
		}
	}
}

func (v customNameserverValidator) getStringValue(ctx context.Context, request resource.ValidateConfigRequest, attr path.Path) types.String {
	var value types.String
	request.Config.GetAttribute(ctx, attr, &value)
	return value
}

func (v customNameserverValidator) getStringDefault(ctx context.Context, request resource.ValidateConfigRequest, attr path.Path) (string, error) {
	defaultReq := defaults.StringRequest{Path: attr}
	defaultResp := defaults.StringResponse{}
	attributes := request.Config.Schema.GetAttributes()

	if attribute, ok := attributes[attr.String()]; ok {
		if attribute.GetType() != types.StringType {
			return "", fmt.Errorf(`Expected attribute "%s" to be a string`, attr.String())
		}

		attribute.(schema.StringAttribute).Default.DefaultString(ctx, defaultReq, &defaultResp)
		return defaultResp.PlanValue.ValueString(), nil
	} else {
		return "", fmt.Errorf(`Attribute "%s" not found`, attr.String())
	}
}
