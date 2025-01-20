// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &BunnynetProvider{}

type BunnynetProvider struct {
	version string
}

type BunnyProviderModel struct {
	ApiKey       types.String `tfsdk:"api_key"`
	ApiUrl       types.String `tfsdk:"api_url"`
	StreamApiUrl types.String `tfsdk:"stream_api_url"`
}

func (p *BunnynetProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bunnynet"
	resp.Version = p.version
}

func (p *BunnynetProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage bunny.net resources with Terraform",
		MarkdownDescription: `
The Bunny Terraform provider allows Terraform users to manage their bunny.net resources.

Before getting started, you will need a bunny.net account and the API key for it.

## Authentication

You can either set the API key directly on the <code>api_key</code> attribute for the provider, or set the <code>BUNNYNET_API_KEY</code> environment variable.

> NOTE: Team member API keys are not supported.
		`,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key. Can also be set using the `BUNNYNET_API_KEY` environment variable.",
				Optional:            true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Optional. The API URL. Defaults to `https://api.bunny.net`.",
				Optional:            true,
			},
			"stream_api_url": schema.StringAttribute{
				MarkdownDescription: "Optional. The Stream API URL. Defaults to `https://video.bunnycdn.com`.",
				Optional:            true,
			},
		},
	}
}

func (p *BunnynetProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BunnyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envApiKey := os.Getenv("BUNNYNET_API_KEY")
	if envApiKey != "" {
		data.ApiKey = types.StringValue(envApiKey)
	}

	envApiUrl := os.Getenv("BUNNYNET_API_URL")
	if envApiUrl != "" {
		data.ApiUrl = types.StringValue(envApiUrl)
	}

	if data.ApiUrl.IsNull() {
		data.ApiUrl = types.StringValue("https://api.bunny.net")
	}

	envStreamApiUrl := os.Getenv("BUNNYNET_STREAM_API_URL")
	if envStreamApiUrl != "" {
		data.StreamApiUrl = types.StringValue(envStreamApiUrl)
	}

	if data.StreamApiUrl.IsNull() {
		data.StreamApiUrl = types.StringValue("https://video.bunnycdn.com")
	}

	userAgent := fmt.Sprintf("Terraform/%s BunnynetProvider/%s", req.TerraformVersion, p.version)
	apiClient := api.NewClient(data.ApiKey.ValueString(), data.ApiUrl.ValueString(), data.StreamApiUrl.ValueString(), userAgent)
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *BunnynetProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewComputeScriptResource,
		NewComputeScriptSecretResource,
		NewComputeScriptVariableResource,
		NewDnsRecordResourceResource,
		NewDnsZoneResourceResource,
		NewPullzoneResource,
		NewPullzoneEdgeruleResource,
		NewPullzoneHostnameResource,
		NewPullzoneOptimizerClassResource,
		NewStorageFileResource,
		NewStorageZoneResource,
		NewStreamCollectionResource,
		NewStreamLibraryResource,
		NewStreamVideoResource,
	}
}

func (p *BunnynetProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDnsRecordDataSource,
		NewDnsZoneDataSource,
		NewRegionDataSource,
		NewVideoLanguageDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BunnynetProvider{
			version: version,
		}
	}
}
