package customtype

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"net/url"
	"strings"
)

var _ basetypes.StringTypable = PullzoneOriginUrlType{}
var _ basetypes.StringValuable = PullzoneOriginUrlValue{}
var _ basetypes.StringValuableWithSemanticEquals = PullzoneOriginUrlValue{}

type PullzoneOriginUrlType struct {
	basetypes.StringType
}

func (t PullzoneOriginUrlType) Equal(o attr.Type) bool {
	other, ok := o.(PullzoneOriginUrlType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t PullzoneOriginUrlType) String() string {
	return "PullzoneOriginUrlType"
}

func (t PullzoneOriginUrlType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := PullzoneOriginUrlValue{
		StringValue: in,
	}

	return value, nil
}

func (t PullzoneOriginUrlType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t PullzoneOriginUrlType) ValueType(ctx context.Context) attr.Value {
	return PullzoneOriginUrlValue{}
}

type PullzoneOriginUrlValue struct {
	basetypes.StringValue
}

func (v PullzoneOriginUrlValue) Equal(o attr.Value) bool {
	other, ok := o.(PullzoneOriginUrlValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v PullzoneOriginUrlValue) Type(ctx context.Context) attr.Type {
	return PullzoneOriginUrlType{}
}

func (v PullzoneOriginUrlValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	newValue, ok := newValuable.(PullzoneOriginUrlValue)

	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	if v.ValueString() == newValue.ValueString() {
		return true, nil
	}

	oldUrl, err := url.Parse(v.ValueString())
	if err != nil {
		diags.AddError("Invalid URL", err.Error())
		return false, diags
	}

	newUrl, err := url.Parse(newValue.ValueString())
	if err != nil {
		diags.AddError("Invalid URL", err.Error())
		return false, diags
	}

	if oldUrl.Scheme != newUrl.Scheme {
		return false, nil
	}

	// @TODO add IDN support
	if !strings.EqualFold(oldUrl.Hostname(), newUrl.Hostname()) {
		return false, nil
	}

	oldPort, newPort := oldUrl.Port(), newUrl.Port()
	if oldPort == "" {
		if oldUrl.Scheme == "http" {
			oldPort = "80"
		}
		if oldUrl.Scheme == "https" {
			oldPort = "443"
		}
	}

	if newPort == "" {
		if newUrl.Scheme == "http" {
			newPort = "80"
		}
		if newUrl.Scheme == "https" {
			newPort = "443"
		}
	}

	if oldPort != newPort {
		return false, nil
	}

	oldPath := strings.TrimRight(oldUrl.Path, "/")
	newPath := strings.TrimRight(newUrl.Path, "/")

	if oldPath != newPath {
		return false, nil
	}

	if oldUrl.RawQuery != newUrl.RawQuery {
		return false, nil
	}

	return true, nil
}
