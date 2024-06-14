package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Routing                            types.Object  `tfsdk:"routing"`
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
	"type":                types.StringType,
	"url":                 types.StringType,
	"storagezone":         types.Int64Type,
	"follow_redirects":    types.BoolType,
	"forward_host_header": types.BoolType,
	"verify_ssl":          types.BoolType,
}

var pullzoneRoutingTypes = map[string]attr.Type{
	"tier": types.StringType,
	"zones": types.SetType{
		ElemType: types.StringType,
	},
	"filters": types.SetType{
		ElemType: types.StringType,
	},
}

var pullzoneOriginTypeMap = map[uint8]string{
	0: "OriginUrl",
	2: "StorageZone",
}

var pullzoneRoutingTierMap = map[uint8]string{
	0: "Standard",
	1: "Volume",
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
	pullzoneRoutingZonesDefault, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("AF"),
		types.StringValue("ASIA"),
		types.StringValue("EU"),
		types.StringValue("SA"),
		types.StringValue("US"),
	})

	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	pullzoneRoutingFiltersDefault, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("all"),
	})

	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

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
					"follow_redirects": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"forward_host_header": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"verify_ssl": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
				},
			},
			"routing": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"tier": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("Standard"),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"zones": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(pullzoneRoutingZonesDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
					},
					"filters": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(pullzoneRoutingFiltersDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
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
	routing := dataTf.Routing.Attributes()

	dataApi := api.Pullzone{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()

	// optimizer
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

	// origin
	dataApi.OriginType = mapValueToKey(pullzoneOriginTypeMap, origin["type"].(types.String).ValueString())
	dataApi.OriginUrl = origin["url"].(types.String).ValueString()
	dataApi.StorageZoneId = origin["storagezone"].(types.Int64).ValueInt64()
	dataApi.AddHostHeader = origin["forward_host_header"].(types.Bool).ValueBool()
	dataApi.VerifyOriginSSL = origin["verify_ssl"].(types.Bool).ValueBool()
	dataApi.FollowRedirects = origin["follow_redirects"].(types.Bool).ValueBool()

	// routing
	dataApi.Type = mapValueToKey(pullzoneRoutingTierMap, routing["tier"].(types.String).ValueString())

	// routing.zones
	{
		zones := routing["zones"].(types.Set).Elements()
		for _, zone := range zones {
			if zone.(types.String).ValueString() == "AF" {
				dataApi.EnableGeoZoneAF = true
			}
			if zone.(types.String).ValueString() == "ASIA" {
				dataApi.EnableGeoZoneASIA = true
			}
			if zone.(types.String).ValueString() == "EU" {
				dataApi.EnableGeoZoneEU = true
			}
			if zone.(types.String).ValueString() == "SA" {
				dataApi.EnableGeoZoneSA = true
			}
			if zone.(types.String).ValueString() == "US" {
				dataApi.EnableGeoZoneUS = true
			}
		}
	}

	// routing.filters
	{
		filters := routing["filters"].(types.Set).Elements()
		values := make([]string, len(filters))
		for i, filter := range filters {
			values[i] = filter.(types.String).ValueString()
		}
		dataApi.RoutingFilters = values
	}

	return dataApi
}

func (r *PullzoneResource) convertApiToModel(dataApi api.Pullzone) (PullzoneResourceModel, diag.Diagnostics) {
	dataTf := PullzoneResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)

	// optimizer
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
	{
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

		originValues["follow_redirects"] = types.BoolValue(dataApi.FollowRedirects)
		originValues["forward_host_header"] = types.BoolValue(dataApi.AddHostHeader)
		originValues["verify_ssl"] = types.BoolValue(dataApi.VerifyOriginSSL)

		origin, diags := types.ObjectValue(pullzoneOriginTypes, originValues)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.Origin = origin
	}

	// routing
	{
		routingValues := map[string]attr.Value{
			"tier": types.StringValue(mapKeyToValue(pullzoneRoutingTierMap, dataApi.Type)),
		}

		// zones
		{
			var zonesValues []attr.Value
			if dataApi.EnableGeoZoneAF {
				zonesValues = append(zonesValues, types.StringValue("AF"))
			}

			if dataApi.EnableGeoZoneASIA {
				zonesValues = append(zonesValues, types.StringValue("ASIA"))
			}

			if dataApi.EnableGeoZoneEU {
				zonesValues = append(zonesValues, types.StringValue("EU"))
			}

			if dataApi.EnableGeoZoneSA {
				zonesValues = append(zonesValues, types.StringValue("SA"))
			}

			if dataApi.EnableGeoZoneUS {
				zonesValues = append(zonesValues, types.StringValue("US"))
			}

			zones, diags := types.SetValue(types.StringType, zonesValues)
			if diags != nil {
				return dataTf, diags
			}

			routingValues["zones"] = zones
		}

		// filters
		{
			filtersValues := make([]attr.Value, len(dataApi.RoutingFilters))
			for i, v := range dataApi.RoutingFilters {
				filtersValues[i] = types.StringValue(v)
			}

			filters, diags := types.SetValue(types.StringType, filtersValues)
			if diags != nil {
				return dataTf, diags
			}

			routingValues["filters"] = filters
		}

		routing, diags := types.ObjectValue(pullzoneRoutingTypes, routingValues)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.Routing = routing
	}

	return dataTf, nil
}
