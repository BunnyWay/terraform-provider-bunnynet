package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	ApiKey       types.String `tfsdk:"api_key"`
	ApiUrl       types.String `tfsdk:"api_url"`
	StreamApiUrl types.String `tfsdk:"stream_api_url"`
}

func (p *BunnyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bunny"
	resp.Version = p.version
}

func (p *BunnyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage bunny.net resources with Terraform",
		MarkdownDescription: `
The Bunny Terraform provider allows Terraform users to manage their bunny.net resources.

Before getting started, you will need a bunny.net account and the API key for it.

> NOTE: Team member API keys are not supported.
		`,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key. Can also be set using the `BUNNY_API_KEY` environment variable.",
				Optional:            true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "The API URL. Defaults to `https://api.bunny.net`.",
				Optional:            true,
			},
			"stream_api_url": schema.StringAttribute{
				MarkdownDescription: "The Stream API URL. Defaults to `https://video.bunnycdn.com`.",
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

	envStreamApiUrl := os.Getenv("BUNNY_STREAM_API_URL")
	if envStreamApiUrl != "" {
		data.StreamApiUrl = types.StringValue(envStreamApiUrl)
	}

	if data.StreamApiUrl.IsNull() {
		data.StreamApiUrl = types.StringValue("https://video.bunnycdn.com")
	}

	userAgent := fmt.Sprintf("Terraform/%s BunnyProvider/%s", req.TerraformVersion, p.version)
	apiClient := api.NewClient(data.ApiKey.ValueString(), data.ApiUrl.ValueString(), data.StreamApiUrl.ValueString(), userAgent)
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient

	if len(streamLibraryLanguageOptions) == 0 {
		languages, err := apiClient.GetVideoLanguages()
		if err != nil {
			resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching video languages", err.Error()))
			return
		}

		streamLibraryLanguageOptions = sliceMap(languages, func(v api.VideoLanguage) string {
			return v.ShortCode
		})
	}
}

func (p *BunnyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
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

func (p *BunnyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRegionDataSource,
		NewVideoLanguageDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BunnyProvider{
			version: version,
		}
	}
}

func convertSetToStringSlice(set types.Set) []string {
	filters := set.Elements()
	values := make([]string, len(filters))
	for i, filter := range filters {
		values[i] = filter.(types.String).ValueString()
	}
	return values
}

func convertStringSliceToSet(values []string) (types.Set, diag.Diagnostics) {
	setValues := make([]attr.Value, len(values))
	for i, v := range values {
		setValues[i] = types.StringValue(v)
	}

	filters, diags := types.SetValue(types.StringType, setValues)
	if diags != nil {
		return types.Set{}, diags
	}

	return filters, nil
}
