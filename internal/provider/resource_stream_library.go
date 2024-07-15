// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"regexp"
	"strconv"
	"strings"
)

var _ resource.Resource = &StreamLibraryResource{}
var _ resource.ResourceWithImportState = &StreamLibraryResource{}

func NewStreamLibraryResource() resource.Resource {
	return &StreamLibraryResource{}
}

type StreamLibraryResource struct {
	client *api.Client
}

type StreamLibraryResourceModel struct {
	Id                                  types.Int64  `tfsdk:"id"`
	Name                                types.String `tfsdk:"name"`
	Pullzone                            types.Int64  `tfsdk:"pullzone"`
	StorageZone                         types.Int64  `tfsdk:"storage_zone"`
	ApiKey                              types.String `tfsdk:"api_key"`
	PlayerLanguage                      types.String `tfsdk:"player_language"`
	PlayerFontFamily                    types.String `tfsdk:"player_font_family"`
	PlayerPrimaryColor                  types.String `tfsdk:"player_primary_color"`
	PlayerControls                      types.Set    `tfsdk:"player_controls"`
	PlayerCustomHead                    types.String `tfsdk:"player_custom_head"`
	PlayerCaptionsFontColor             types.String `tfsdk:"player_captions_font_color"`
	PlayerCaptionsFontSize              types.Int64  `tfsdk:"player_captions_font_size"`
	PlayerCaptionsBackgroundColor       types.String `tfsdk:"player_captions_background_color"`
	PlayerWatchtimeHeatmapEnabled       types.Bool   `tfsdk:"player_watchtime_heatmap_enabled"`
	VastTagUrl                          types.String `tfsdk:"vast_tag_url"`
	OriginalFilesKeep                   types.Bool   `tfsdk:"original_files_keep"`
	EarlyPlayEnabled                    types.Bool   `tfsdk:"early_play_enabled"`
	ContentTaggingEnabled               types.Bool   `tfsdk:"content_tagging_enabled"`
	Mp4FallbackEnabled                  types.Bool   `tfsdk:"mp4_fallback_enabled"`
	Resolutions                         types.Set    `tfsdk:"resolutions"`
	Bitrate240p                         types.Int64  `tfsdk:"bitrate_240p"`
	Bitrate360p                         types.Int64  `tfsdk:"bitrate_360p"`
	Bitrate480p                         types.Int64  `tfsdk:"bitrate_480p"`
	Bitrate720p                         types.Int64  `tfsdk:"bitrate_720p"`
	Bitrate1080p                        types.Int64  `tfsdk:"bitrate_1080p"`
	Bitrate1440p                        types.Int64  `tfsdk:"bitrate_1440p"`
	Bitrate2160p                        types.Int64  `tfsdk:"bitrate_2160p"`
	WatermarkPositionLeft               types.Int64  `tfsdk:"watermark_position_left"`
	WatermarkPositionTop                types.Int64  `tfsdk:"watermark_position_top"`
	WatermarkWidth                      types.Int64  `tfsdk:"watermark_width"`
	WatermarkHeight                     types.Int64  `tfsdk:"watermark_height"`
	TranscribingEnabled                 types.Bool   `tfsdk:"transcribing_enabled"`
	TranscribingSmartTitleEnabled       types.Bool   `tfsdk:"transcribing_smart_title_enabled"`
	TranscribingSmartDescriptionEnabled types.Bool   `tfsdk:"transcribing_smart_description_enabled"`
	TranscribingLanguages               types.Set    `tfsdk:"transcribing_languages"`
	DirectPlayEnabled                   types.Bool   `tfsdk:"direct_play_enabled"`
	ReferersAllowed                     types.Set    `tfsdk:"referers_allowed"`
	ReferersBlocked                     types.Set    `tfsdk:"referers_blocked"`
	DirectUrlFileAccessBlocked          types.Bool   `tfsdk:"direct_url_file_access_blocked"`
	ViewTokenAuthenticationRequired     types.Bool   `tfsdk:"view_token_authentication_required"`
	CdnTokenAuthenticationRequired      types.Bool   `tfsdk:"cdn_token_authentication_required"`
	DrmMediacageBasicEnabled            types.Bool   `tfsdk:"drm_mediacage_basic_enabled"`
	WebhookUrl                          types.String `tfsdk:"webhook_url"`
}

var streamLibraryFontFamilyOptions = []string{"arial", "inter", "lato", "oswald", "raleway", "roboto", "rubik", "ubuntu"}
var streamLibraryPlayerControlsOptions = []string{"airplay", "captions", "chromecast", "current-time", "duration", "fast-forward", "fullscreen", "mute", "pip", "play", "play-large", "progress", "rewind", "settings", "volume"}

func (r *StreamLibraryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream_library"
}

func (r *StreamLibraryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	streamLibraryPlayerControlsDefault := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("play-large"),
		types.StringValue("play"),
		types.StringValue("progress"),
		types.StringValue("current-time"),
		types.StringValue("mute"),
		types.StringValue("volume"),
		types.StringValue("captions"),
		types.StringValue("settings"),
		types.StringValue("airplay"),
		types.StringValue("pip"),
		types.StringValue("fullscreen"),
	})

	streamLibraryResolutionsDefault := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("240p"),
		types.StringValue("360p"),
		types.StringValue("480p"),
		types.StringValue("720p"),
		types.StringValue("1080p"),
	})

	resp.Schema = schema.Schema{
		MarkdownDescription: "Stream Library",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the Video Library.",
			},
			"pullzone": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the connected underlying pull zone",
			},
			"storage_zone": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the connected underlying storage zone",
			},
			"api_key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The API key used for authenticating with the video library",
			},
			"player_language": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("en"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The UI language of the player",
			},
			"player_font_family": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("rubik"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(streamLibraryFontFamilyOptions...),
				},
				MarkdownDescription: generateMarkdownSliceOptions(streamLibraryFontFamilyOptions),
			},
			"player_primary_color": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("#ff7755"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^#([0-9a-fA-F]{1,6})$"), "Invalid hex color"),
				},
				MarkdownDescription: "The key color of the player.",
			},
			"player_controls": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(streamLibraryPlayerControlsDefault),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(streamLibraryPlayerControlsOptions...),
					),
				},
				MarkdownDescription: generateMarkdownSliceOptions(streamLibraryPlayerControlsOptions),
			},
			"player_custom_head": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The custom HTMl that is added into the head of the HTML player.",
			},
			"player_captions_font_color": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("#fff"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^#([0-9a-fA-F]{1,6})$"), "Invalid hex color"),
				},
				Description: "The captions display font color",
			},
			"player_captions_font_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(20),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The captions display font size",
			},
			"player_captions_background_color": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("#000"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^#([0-9a-fA-F]{1,6})$"), "Invalid hex color"),
				},
				Description: "The captions display background color",
			},
			"player_watchtime_heatmap_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the video watch heatmap should be displayed in the player.",
			},
			"vast_tag_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The URL of the VAST tag endpoint for advertising configuration",
			},
			"original_files_keep": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the original video files should be stored after encoding",
			},
			"early_play_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the Early-Play feature is enabled",
			},
			"content_tagging_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if content tagging should be enabled for this library.",
			},
			"mp4_fallback_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the MP4 fallback feature is enabled",
			},
			"resolutions": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(streamLibraryResolutionsDefault),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "The comma separated list of enabled resolutions",
			},
			"bitrate_240p": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(600),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
				Description: "The bitrate used for encoding 240p videos",
			},
			"bitrate_360p": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(800),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
				Description: "The bitrate used for encoding 360p videos",
			},
			"bitrate_480p": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1400),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
				Description: "The bitrate used for encoding 480p videos",
			},
			"bitrate_720p": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(2800),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
				Description: "The bitrate used for encoding 720p videos",
			},
			"bitrate_1080p": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(5000),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
				Description: "The bitrate used for encoding 1080p videos",
			},
			"bitrate_1440p": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(8000),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
				Description: "The bitrate used for encoding 1440p videos",
			},
			"bitrate_2160p": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(25000),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 4294967295),
				},
				Description: "The bitrate used for encoding 2160p videos",
			},
			"watermark_position_left": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The left offset of the watermark position (in %)",
			},
			"watermark_position_top": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The top offset of the watermark position (in %)",
			},
			"watermark_width": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The width of the watermark (in %)",
			},
			"watermark_height": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The height of the watermark (in %)",
			},
			"transcribing_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the automatic audio transcribing is currently enabled for this zone.",
			},
			"transcribing_smart_title_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if automatic transcribing title generation is currently enabled.",
			},
			"transcribing_smart_description_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if automatic transcribing description generation is currently enabled.",
			},
			"transcribing_languages": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				MarkdownDescription: "The list of languages that the captions will be automatically transcribed to.",
			},
			"direct_play_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines direct play URLs are enabled for the library",
			},
			"referers_allowed": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "The list of allowed referrer domains allowed to access the library",
			},
			"referers_blocked": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "The list of blocked referrer domains blocked from accessing the library",
			},
			"direct_url_file_access_blocked": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the requests without a referrer are blocked",
			},
			"view_token_authentication_required": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the player token authentication is enabled",
			},
			"cdn_token_authentication_required": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"drm_mediacage_basic_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the MediaCage basic DRM is enabled",
			},
			"webhook_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The webhook URL of the video library",
			},
		},
	}
}

func (r *StreamLibraryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *StreamLibraryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf StreamLibraryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateStreamLibrary(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create stream library", err.Error())
		return
	}

	tflog.Trace(ctx, "created video library "+dataApi.Name)
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamLibraryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StreamLibraryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetStreamLibrary(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching stream library", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamLibraryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StreamLibraryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateStreamLibrary(dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating stream library", err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated video library %d", dataApi.Id))

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamLibraryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StreamLibraryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStreamLibrary(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting stream library", err.Error()))
	}
}

func (r *StreamLibraryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error converting ID to integer", err.Error()))
		return
	}

	dataApi, err := r.client.GetStreamLibrary(id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching stream library", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamLibraryResource) convertModelToApi(ctx context.Context, dataTf StreamLibraryResourceModel) api.StreamLibrary {
	dataApi := api.StreamLibrary{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()

	// player
	{
		dataApi.UILanguage = dataTf.PlayerLanguage.ValueString()
		dataApi.FontFamily = dataTf.PlayerFontFamily.ValueString()
		dataApi.PlayerKeyColor = dataTf.PlayerPrimaryColor.ValueString()
		dataApi.Controls = strings.Join(convertSetToStringSlice(dataTf.PlayerControls), ",")
		dataApi.CaptionsFontColor = dataTf.PlayerCaptionsFontColor.ValueString()
		dataApi.CaptionsFontSize = uint16(dataTf.PlayerCaptionsFontSize.ValueInt64())
		dataApi.CaptionsBackground = dataTf.PlayerCaptionsBackgroundColor.ValueString()
		dataApi.ShowHeatmap = dataTf.PlayerWatchtimeHeatmapEnabled.ValueBool()

		customHtml := dataTf.PlayerCustomHead.ValueString()
		dataApi.CustomHTML = &customHtml
	}

	// advertising
	dataApi.VastTagUrl = dataTf.VastTagUrl.ValueString()

	// encoding
	{
		dataApi.KeepOriginalFiles = dataTf.OriginalFilesKeep.ValueBool()
		dataApi.AllowEarlyPlay = dataTf.EarlyPlayEnabled.ValueBool()
		dataApi.EnableContentTagging = dataTf.ContentTaggingEnabled.ValueBool()
		dataApi.EnableMP4Fallback = dataTf.Mp4FallbackEnabled.ValueBool()
		dataApi.EnabledResolutions = strings.Join(convertSetToStringSlice(dataTf.Resolutions), ",")
		dataApi.Bitrate240P = uint32(dataTf.Bitrate240p.ValueInt64())
		dataApi.Bitrate360P = uint32(dataTf.Bitrate360p.ValueInt64())
		dataApi.Bitrate480P = uint32(dataTf.Bitrate480p.ValueInt64())
		dataApi.Bitrate720P = uint32(dataTf.Bitrate720p.ValueInt64())
		dataApi.Bitrate1080P = uint32(dataTf.Bitrate1080p.ValueInt64())
		dataApi.Bitrate1440P = uint32(dataTf.Bitrate1440p.ValueInt64())
		dataApi.Bitrate2160P = uint32(dataTf.Bitrate2160p.ValueInt64())
		dataApi.WatermarkPositionLeft = uint8(dataTf.WatermarkPositionLeft.ValueInt64())
		dataApi.WatermarkPositionTop = uint8(dataTf.WatermarkPositionTop.ValueInt64())
		dataApi.WatermarkWidth = uint16(dataTf.WatermarkWidth.ValueInt64())
		dataApi.WatermarkHeight = uint16(dataTf.WatermarkHeight.ValueInt64())
	}

	// transcribing
	dataApi.EnableTranscribing = dataTf.TranscribingEnabled.ValueBool()
	dataApi.EnableTranscribingTitleGeneration = dataTf.TranscribingSmartTitleEnabled.ValueBool()
	dataApi.EnableTranscribingDescriptionGeneration = dataTf.TranscribingSmartDescriptionEnabled.ValueBool()
	dataApi.TranscribingCaptionLanguages = convertSetToStringSlice(dataTf.TranscribingLanguages)

	// security
	{
		dataApi.AllowDirectPlay = dataTf.DirectPlayEnabled.ValueBool()
		dataApi.AllowedReferrers = convertSetToStringSlice(dataTf.ReferersAllowed)
		dataApi.BlockedReferrers = convertSetToStringSlice(dataTf.ReferersBlocked)
		dataApi.BlockNoneReferrer = dataTf.DirectUrlFileAccessBlocked.ValueBool()
		dataApi.PlayerTokenAuthenticationEnabled = dataTf.ViewTokenAuthenticationRequired.ValueBool()
		dataApi.EnableTokenAuthentication = dataTf.CdnTokenAuthenticationRequired.ValueBool()
		dataApi.EnableDRM = dataTf.DrmMediacageBasicEnabled.ValueBool()
	}

	// api
	webhookUrl := dataTf.WebhookUrl.ValueString()
	dataApi.WebhookUrl = &webhookUrl

	return dataApi
}

func (r *StreamLibraryResource) convertApiToModel(dataApi api.StreamLibrary) (StreamLibraryResourceModel, diag.Diagnostics) {
	dataTf := StreamLibraryResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Pullzone = types.Int64Value(dataApi.PullZoneId)
	dataTf.StorageZone = types.Int64Value(dataApi.StorageZoneId)
	dataTf.ApiKey = types.StringValue(dataApi.ApiKey)

	// player
	{
		dataTf.PlayerLanguage = types.StringValue(dataApi.UILanguage)
		dataTf.PlayerFontFamily = types.StringValue(dataApi.FontFamily)
		dataTf.PlayerPrimaryColor = types.StringValue(dataApi.PlayerKeyColor)
		dataTf.PlayerCaptionsFontColor = types.StringValue(dataApi.CaptionsFontColor)
		dataTf.PlayerCaptionsFontSize = types.Int64Value(int64(dataApi.CaptionsFontSize))
		dataTf.PlayerCaptionsBackgroundColor = types.StringValue(dataApi.CaptionsBackground)
		dataTf.PlayerWatchtimeHeatmapEnabled = types.BoolValue(dataApi.ShowHeatmap)

		// controls
		{
			if len(dataApi.Controls) == 0 {
				dataTf.PlayerControls = types.SetValueMust(types.StringType, []attr.Value{})
			} else {
				var values []attr.Value
				for _, control := range strings.Split(dataApi.Controls, ",") {
					values = append(values, types.StringValue(control))
				}

				controlSet, diags := types.SetValue(types.StringType, values)
				if diags != nil {
					return dataTf, diags
				}

				dataTf.PlayerControls = controlSet
			}
		}

		if dataApi.CustomHTML == nil {
			dataTf.PlayerCustomHead = types.StringValue("")
		} else {
			dataTf.PlayerCustomHead = types.StringValue(*dataApi.CustomHTML)
		}
	}

	// advertising
	dataTf.VastTagUrl = types.StringValue(dataApi.VastTagUrl)

	// encoding
	{
		dataTf.OriginalFilesKeep = types.BoolValue(dataApi.KeepOriginalFiles)
		dataTf.EarlyPlayEnabled = types.BoolValue(dataApi.AllowEarlyPlay)
		dataTf.ContentTaggingEnabled = types.BoolValue(dataApi.EnableContentTagging)
		dataTf.Mp4FallbackEnabled = types.BoolValue(dataApi.EnableMP4Fallback)
		dataTf.Bitrate240p = types.Int64Value(int64(dataApi.Bitrate240P))
		dataTf.Bitrate360p = types.Int64Value(int64(dataApi.Bitrate360P))
		dataTf.Bitrate480p = types.Int64Value(int64(dataApi.Bitrate480P))
		dataTf.Bitrate720p = types.Int64Value(int64(dataApi.Bitrate720P))
		dataTf.Bitrate1080p = types.Int64Value(int64(dataApi.Bitrate1080P))
		dataTf.Bitrate1440p = types.Int64Value(int64(dataApi.Bitrate1440P))
		dataTf.Bitrate2160p = types.Int64Value(int64(dataApi.Bitrate2160P))

		// resolution
		{
			if len(dataApi.EnabledResolutions) == 0 {
				dataTf.Resolutions = types.SetValueMust(types.StringType, []attr.Value{})
			} else {
				var values []attr.Value
				for _, resolution := range strings.Split(dataApi.EnabledResolutions, ",") {
					values = append(values, types.StringValue(resolution))
				}

				resolutionSet, diags := types.SetValue(types.StringType, values)
				if diags != nil {
					return dataTf, diags
				}

				dataTf.Resolutions = resolutionSet
			}
		}

		dataTf.WatermarkPositionLeft = types.Int64Value(int64(dataApi.WatermarkPositionLeft))
		dataTf.WatermarkPositionTop = types.Int64Value(int64(dataApi.WatermarkPositionTop))
		dataTf.WatermarkWidth = types.Int64Value(int64(dataApi.WatermarkWidth))
		dataTf.WatermarkHeight = types.Int64Value(int64(dataApi.WatermarkHeight))
	}

	// transcribing
	{
		transcribingLanguages, diags := convertStringSliceToSet(dataApi.TranscribingCaptionLanguages)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.TranscribingEnabled = types.BoolValue(dataApi.EnableTranscribing)
		dataTf.TranscribingSmartTitleEnabled = types.BoolValue(dataApi.EnableTranscribingTitleGeneration)
		dataTf.TranscribingSmartDescriptionEnabled = types.BoolValue(dataApi.EnableTranscribingDescriptionGeneration)
		dataTf.TranscribingLanguages = transcribingLanguages
	}

	// security
	{
		referersAllowed, diags := convertStringSliceToSet(dataApi.AllowedReferrers)
		if diags != nil {
			return dataTf, diags
		}

		referersBlocked, diags := convertStringSliceToSet(dataApi.BlockedReferrers)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.DirectPlayEnabled = types.BoolValue(dataApi.AllowDirectPlay)
		dataTf.ReferersAllowed = referersAllowed
		dataTf.ReferersBlocked = referersBlocked
		dataTf.DirectUrlFileAccessBlocked = types.BoolValue(dataApi.BlockNoneReferrer)
		dataTf.ViewTokenAuthenticationRequired = types.BoolValue(dataApi.PlayerTokenAuthenticationEnabled)
		dataTf.CdnTokenAuthenticationRequired = types.BoolValue(dataApi.EnableTokenAuthentication)
		dataTf.DrmMediacageBasicEnabled = types.BoolValue(dataApi.EnableDRM)
	}

	// api
	if dataApi.WebhookUrl == nil {
		dataTf.WebhookUrl = types.StringValue("")
	} else {
		dataTf.WebhookUrl = types.StringValue(*dataApi.WebhookUrl)
	}

	return dataTf, nil
}
