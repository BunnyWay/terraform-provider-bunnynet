// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSecretAttributesAreSensitive(t *testing.T) {
	cases := []struct {
		resource  func() resource.Resource
		attribute string
	}{
		{NewPullzoneResource, "s3_auth_secret"},
		{NewPullzoneResource, "log_forward_token"},
		{NewPullzoneResource, "token_auth_key"},
		{NewPullzoneHostnameResource, "certificate_key"},
		{NewComputeScriptResource, "deployment_key"},
	}

	for _, c := range cases {
		r := c.resource()
		var meta resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "bunnynet"}, &meta)

		var resp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("%s: schema diagnostics: %v", meta.TypeName, resp.Diagnostics)
		}

		attr, ok := resp.Schema.Attributes[c.attribute]
		if !ok {
			t.Fatalf("%s: attribute %q not found", meta.TypeName, c.attribute)
		}
		if !attr.IsSensitive() {
			t.Errorf("%s: attribute %q must be marked as sensitive", meta.TypeName, c.attribute)
		}
	}
}

func TestPullzoneDataSourceSecretAttributesAreSensitive(t *testing.T) {
	// The data source schema is derived from the resource schema, so a secret
	// marked sensitive on the resource must stay sensitive on the data source.
	d := NewPullzoneDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("data source pullzone: schema diagnostics: %v", resp.Diagnostics)
	}

	for _, name := range []string{"s3_auth_secret", "log_forward_token", "token_auth_key"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("data source pullzone: attribute %q not found", name)
		}
		if !attr.IsSensitive() {
			t.Errorf("data source pullzone: attribute %q must be marked as sensitive", name)
		}
	}
}

func TestResourceAttrToDatasourceAttrPreservesSensitive(t *testing.T) {
	for _, sensitive := range []bool{true, false} {
		cases := map[string]rschema.Attribute{
			"bool":    rschema.BoolAttribute{Sensitive: sensitive},
			"float64": rschema.Float64Attribute{Sensitive: sensitive},
			"int64":   rschema.Int64Attribute{Sensitive: sensitive},
			"set":     rschema.SetAttribute{ElementType: types.StringType, Sensitive: sensitive},
			"string":  rschema.StringAttribute{Sensitive: sensitive},
		}

		for name, ra := range cases {
			da := resourceAttrToDatasourceAttr(ra)
			if da.IsSensitive() != sensitive {
				t.Errorf("%s: expected sensitive=%v, got %v", name, sensitive, da.IsSensitive())
			}
			if !da.IsComputed() {
				t.Errorf("%s: data source attribute must be computed", name)
			}
		}
	}
}
