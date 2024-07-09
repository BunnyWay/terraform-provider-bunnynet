package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &VideoLanguageDataSource{}

func NewVideoLanguageDataSource() datasource.DataSource {
	return &VideoLanguageDataSource{}
}

type VideoLanguageDataSource struct {
	client *api.Client
}

type VideoLanguageDataSourceModel struct {
	Code                     types.String `tfsdk:"code"`
	Name                     types.String `tfsdk:"name"`
	SupportPlayerTranslation types.Bool   `tfsdk:"support_player_translation"`
	SupportTranscribing      types.Bool   `tfsdk:"support_transcribing"`
	TranscribingAccuracy     types.Int64  `tfsdk:"transcribing_accuracy"`
}

func (d *VideoLanguageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_video_language"
}

func (d *VideoLanguageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Video Language",
		Attributes: map[string]schema.Attribute{
			"code": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"support_player_translation": schema.BoolAttribute{
				Computed: true,
			},
			"support_transcribing": schema.BoolAttribute{
				Computed: true,
			},
			"transcribing_accuracy": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *VideoLanguageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *VideoLanguageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var dataTf VideoLanguageDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := d.client.GetVideoLanguage(dataTf.Code.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read video languages, got error: %s", err))
		return
	}

	dataTf.Code = types.StringValue(dataApi.ShortCode)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.SupportPlayerTranslation = types.BoolValue(dataApi.SupportPlayerTranslation)
	dataTf.SupportTranscribing = types.BoolValue(dataApi.SupportTranscribing)
	dataTf.TranscribingAccuracy = types.Int64Value(dataApi.TranscribingAccuracy)

	tflog.Trace(ctx, "read video dataApi "+dataApi.ShortCode)
	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}
