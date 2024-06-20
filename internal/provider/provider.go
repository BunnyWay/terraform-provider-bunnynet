package provider

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &BunnyProvider{}

type BunnyProvider struct {
	version string
}

type BunnyProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
	ApiUrl types.String `tfsdk:"api_url"`
}

func (p *BunnyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bunny"
	resp.Version = p.version
}

func (p *BunnyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key",
				Optional:            true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "API URL",
				Optional:            true,
			},
		},
	}
}

func (p *BunnyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BunnyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envApiKey := os.Getenv("BUNNY_API_KEY")
	if envApiKey != "" {
		data.ApiKey = types.StringValue(envApiKey)
	}

	envApiUrl := os.Getenv("BUNNY_API_URL")
	if envApiUrl != "" {
		data.ApiUrl = types.StringValue(envApiUrl)
	}

	if data.ApiUrl.IsNull() {
		data.ApiUrl = types.StringValue("https://api.bunny.net")
	}

	apiClient := api.NewClient(data.ApiKey.ValueString(), data.ApiUrl.ValueString())
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *BunnyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDnsRecordResourceResource,
		NewDnsZoneResourceResource,
		NewPullzoneResource,
		NewPullzoneEdgeruleResource,
		NewPullzoneHostnameResource,
		NewPullzoneOptimizerClassResource,
		NewStorageZoneResource,
	}
}

func (p *BunnyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRegionDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BunnyProvider{
			version: version,
		}
	}
}
