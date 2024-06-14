package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
)

var _ resource.Resource = &PullzoneResource{}
var _ resource.ResourceWithImportState = &PullzoneResource{}

func NewPullzoneResource() resource.Resource {
	return &PullzoneResource{}
}

type PullzoneResource struct {
	client        *api.Client
	pullzoneMutex *pullzoneMutex
}

type PullzoneResourceModel struct {
	Id                                 types.Int64   `tfsdk:"id"`
	Name                               types.String  `tfsdk:"name"`
	Origin                             types.Object  `tfsdk:"origin"`
	OptimizerEnabled                   types.Bool    `tfsdk:"optimizer_enabled"`
	OptimizerMinifyCss                 types.Bool    `tfsdk:"optimizer_minify_css"`
	OptimizerMinifyJs                  types.Bool    `tfsdk:"optimizer_minify_js"`
	OptimizerWebp                      types.Bool    `tfsdk:"optimizer_webp"`
	OptimizerClassesForce              types.Bool    `tfsdk:"optimizer_classes_force"`
	OptimizerDynamicImageApi           types.Bool    `tfsdk:"optimizer_dynamic_image_api"`
	OptimizerSmartImage                types.Bool    `tfsdk:"optimizer_smartimage"`
	OptimizerSmartImageDesktopMaxwidth types.Int64   `tfsdk:"optimizer_smartimage_desktop_maxwidth"`
	OptimizerSmartImageDesktopQuality  types.Int64   `tfsdk:"optimizer_smartimage_desktop_quality"`
	OptimizerSmartImageMobileMaxwidth  types.Int64   `tfsdk:"optimizer_smartimage_mobile_maxwidth"`
	OptimizerSmartImageMobileQuality   types.Int64   `tfsdk:"optimizer_smartimage_mobile_quality"`
	OptimizerWatermark                 types.Bool    `tfsdk:"optimizer_watermark"`
	OptimizerWatermarkUrl              types.String  `tfsdk:"optimizer_watermark_url"`
	OptimizerWatermarkPosition         types.String  `tfsdk:"optimizer_watermark_position"`
	OptimizerWatermarkBorderoffset     types.Float64 `tfsdk:"optimizer_watermark_borderoffset"`
	OptimizerWatermarkMinsize          types.Int64   `tfsdk:"optimizer_watermark_minsize"`
}

var pullzoneOriginTypes = map[string]attr.Type{
	"type":        types.StringType,
	"url":         types.StringType,
	"storagezone": types.Int64Type,
}

var pullzoneOriginTypeMap = map[uint8]string{
	0: "OriginUrl",
	2: "StorageZone",
}

var pullzoneOptimizerWatermarkPositionMap = map[uint8]string{
	0: "BottomLeft",
	1: "BottomRight",
	2: "TopLeft",
	3: "TopRight",
	4: "Center",
	5: "CenterStretch",
}

func (r *PullzoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone"
}

func (r *PullzoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pullzone",

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
			},
			"optimizer_classes_force": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_dynamic_image_api": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_minify_css": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_minify_js": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_smartimage": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_smartimage_desktop_maxwidth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1600),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_smartimage_desktop_quality": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(85),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_smartimage_mobile_maxwidth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(800),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_smartimage_mobile_quality": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(70),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_watermark": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_watermark_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_watermark_position": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("BottomLeft"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_watermark_borderoffset": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(3.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_watermark_minsize": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(300),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_webp": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"origin": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"url": schema.StringAttribute{
						Optional: true,
					},
					"storagezone": schema.Int64Attribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (r *PullzoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzone(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create pullzone", err.Error())
		return
	}

	tflog.Trace(ctx, "created pullzone "+dataApi.Name)
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetPullzone(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching pullzone", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.Id.ValueInt64()
	pzMutex.Lock(pullzoneId)
	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdatePullzone(dataApi)
	pzMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating pullzone", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.Id.ValueInt64()
	pzMutex.Lock(pullzoneId)
	err := r.client.DeletePullzone(pullzoneId)
	pzMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting pullzone", err.Error()))
	}
}

func (r *PullzoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error converting ID to integer", err.Error()))
		return
	}

	dataApi, err := r.client.GetPullzone(id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching pullzone", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneResource) convertModelToApi(ctx context.Context, dataTf PullzoneResourceModel) api.Pullzone {
	origin := dataTf.Origin.Attributes()

	dataApi := api.Pullzone{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.OriginType = mapValueToKey(pullzoneOriginTypeMap, origin["type"].(types.String).ValueString())
	dataApi.OriginUrl = origin["url"].(types.String).ValueString()
	dataApi.StorageZoneId = origin["storagezone"].(types.Int64).ValueInt64()

	dataApi.OptimizerEnabled = dataTf.OptimizerEnabled.ValueBool()
	dataApi.OptimizerMinifyCss = dataTf.OptimizerMinifyCss.ValueBool()
	dataApi.OptimizerMinifyJs = dataTf.OptimizerMinifyJs.ValueBool()
	dataApi.OptimizerWebp = dataTf.OptimizerWebp.ValueBool()
	dataApi.OptimizerForceClasses = dataTf.OptimizerClassesForce.ValueBool()
	dataApi.OptimizerImageOptimization = dataTf.OptimizerDynamicImageApi.ValueBool()
	dataApi.OptimizerAutomaticOptimizationEnabled = dataTf.OptimizerSmartImage.ValueBool()
	dataApi.OptimizerDesktopMaxWidth = uint64(dataTf.OptimizerSmartImageDesktopMaxwidth.ValueInt64())
	dataApi.OptimizerImageQuality = uint8(dataTf.OptimizerSmartImageDesktopQuality.ValueInt64())
	dataApi.OptimizerMobileMaxWidth = uint64(dataTf.OptimizerSmartImageMobileMaxwidth.ValueInt64())
	dataApi.OptimizerMobileImageQuality = uint8(dataTf.OptimizerSmartImageMobileQuality.ValueInt64())
	dataApi.OptimizerWatermarkEnabled = dataTf.OptimizerWatermark.ValueBool()
	dataApi.OptimizerWatermarkUrl = dataTf.OptimizerWatermarkUrl.ValueString()
	dataApi.OptimizerWatermarkPosition = mapValueToKey(pullzoneOptimizerWatermarkPositionMap, dataTf.OptimizerWatermarkPosition.ValueString())
	dataApi.OptimizerWatermarkOffset = dataTf.OptimizerWatermarkBorderoffset.ValueFloat64()
	dataApi.OptimizerWatermarkMinImageSize = uint64(dataTf.OptimizerWatermarkMinsize.ValueInt64())

	return dataApi
}

func (r *PullzoneResource) convertApiToModel(dataApi api.Pullzone) (PullzoneResourceModel, diag.Diagnostics) {
	dataTf := PullzoneResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.OptimizerEnabled = types.BoolValue(dataApi.OptimizerEnabled)
	dataTf.OptimizerMinifyCss = types.BoolValue(dataApi.OptimizerMinifyCss)
	dataTf.OptimizerMinifyJs = types.BoolValue(dataApi.OptimizerMinifyJs)
	dataTf.OptimizerWebp = types.BoolValue(dataApi.OptimizerWebp)
	dataTf.OptimizerClassesForce = types.BoolValue(dataApi.OptimizerForceClasses)
	dataTf.OptimizerDynamicImageApi = types.BoolValue(dataApi.OptimizerImageOptimization)
	dataTf.OptimizerSmartImage = types.BoolValue(dataApi.OptimizerAutomaticOptimizationEnabled)
	dataTf.OptimizerSmartImageDesktopMaxwidth = types.Int64Value(int64(dataApi.OptimizerDesktopMaxWidth))
	dataTf.OptimizerSmartImageDesktopQuality = types.Int64Value(int64(dataApi.OptimizerImageQuality))
	dataTf.OptimizerSmartImageMobileMaxwidth = types.Int64Value(int64(dataApi.OptimizerMobileMaxWidth))
	dataTf.OptimizerSmartImageMobileQuality = types.Int64Value(int64(dataApi.OptimizerMobileImageQuality))
	dataTf.OptimizerWatermark = types.BoolValue(dataApi.OptimizerWatermarkEnabled)
	dataTf.OptimizerWatermarkUrl = types.StringValue(dataApi.OptimizerWatermarkUrl)
	dataTf.OptimizerWatermarkPosition = types.StringValue(mapKeyToValue(pullzoneOptimizerWatermarkPositionMap, dataApi.OptimizerWatermarkPosition))
	dataTf.OptimizerWatermarkBorderoffset = types.Float64Value(dataApi.OptimizerWatermarkOffset)
	dataTf.OptimizerWatermarkMinsize = types.Int64Value(int64(dataApi.OptimizerWatermarkMinImageSize))

	// origin
	originValues := map[string]attr.Value{
		"type": types.StringValue(mapKeyToValue(pullzoneOriginTypeMap, dataApi.OriginType)),
	}

	if dataApi.OriginUrl == "" {
		originValues["url"] = types.StringNull()
	} else {
		originValues["url"] = types.StringValue(dataApi.OriginUrl)
	}

	if dataApi.StorageZoneId == 0 || dataApi.StorageZoneId == -1 {
		originValues["storagezone"] = types.Int64Null()
	} else {
		originValues["storagezone"] = types.Int64Value(dataApi.StorageZoneId)
	}

	origin, diags := types.ObjectValue(pullzoneOriginTypes, originValues)
	if diags != nil {
		return dataTf, diags
	}

	dataTf.Origin = origin

	return dataTf, nil
}
