package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	"golang.org/x/exp/slices"
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
	CacheEnabled                       types.Bool    `tfsdk:"cache_enabled"`
	CacheExpirationTime                types.Int64   `tfsdk:"cache_expiration_time"`
	CacheExpirationTimeBrowser         types.Int64   `tfsdk:"cache_expiration_time_browser"`
	SortQueryString                    types.Bool    `tfsdk:"sort_querystring"`
	CacheErrors                        types.Bool    `tfsdk:"cache_errors"`
	CacheVary                          types.Set     `tfsdk:"cache_vary"`
	CacheVaryQueryStringValues         types.Set     `tfsdk:"cache_vary_querystring"`
	CacheVaryCookieValues              types.Set     `tfsdk:"cache_vary_cookie"`
	StripCookies                       types.Bool    `tfsdk:"strip_cookies"`
	CacheChunked                       types.Bool    `tfsdk:"cache_chunked"`
	CacheStale                         types.Set     `tfsdk:"cache_stale"`
	PermacacheStoragezone              types.Int64   `tfsdk:"permacache_storagezone"`
	OriginShieldEnabled                types.Bool    `tfsdk:"originshield_enabled"`
	OriginShieldConcurrencyLimit       types.Bool    `tfsdk:"originshield_concurrency_limit"`
	OriginShieldConcurrencyRequests    types.Int64   `tfsdk:"originshield_concurrency_requests"`
	OriginShieldQueueRequests          types.Int64   `tfsdk:"originshield_queue_requests"`
	OriginShieldQueueWait              types.Int64   `tfsdk:"originshield_queue_wait"`
	RequestCoalescingEnabled           types.Bool    `tfsdk:"request_coalescing_enabled"`
	RequestCoalescingTimeout           types.Int64   `tfsdk:"request_coalescing_timeout"`
	CorsEnabled                        types.Bool    `tfsdk:"cors_enabled"`
	CorsExtensions                     types.Set     `tfsdk:"cors_extensions"`
	Origin                             types.Object  `tfsdk:"origin"`
	Routing                            types.Object  `tfsdk:"routing"`
	LimitDownloadSpeed                 types.Float64 `tfsdk:"limit_download_speed"`
	LimitRequests                      types.Int64   `tfsdk:"limit_requests"`
	LimitAfter                         types.Float64 `tfsdk:"limit_after"`
	LimitBurst                         types.Int64   `tfsdk:"limit_burst"`
	LimitConnections                   types.Int64   `tfsdk:"limit_connections"`
	LimitBandwidth                     types.Int64   `tfsdk:"limit_bandwidth"`
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
	SafehopEnabled                     types.Bool    `tfsdk:"safehop_enabled"`
	SafehopRetryCount                  types.Int64   `tfsdk:"safehop_retry_count"`
	SafehopRetryDelay                  types.Int64   `tfsdk:"safehop_retry_delay"`
	SafehopRetryReasons                types.Set     `tfsdk:"safehop_retry_reasons"`
	SafehopConnectionTimeout           types.Int64   `tfsdk:"safehop_connection_timeout"`
	SafehopResponseTimeout             types.Int64   `tfsdk:"safehop_response_timeout"`
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
	pullzoneCorsExtensionsDefault, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("css"),
		types.StringValue("eot"),
		types.StringValue("ttf"),
		types.StringValue("woff"),
		types.StringValue("woff2"),
	})

	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

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

	pullzoneSafehopRetryReasonsDefault, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("connectionTimeout"),
		types.StringValue("responseTimeout"),
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
			"cache_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_expiration_time": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(-1, 31919000), // -1 to 1y
				},
			},
			"cache_expiration_time_browser": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(-1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(-1, 31919000), // -1 to 1y
				},
			},
			"sort_querystring": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_errors": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_vary": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_vary_querystring": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_vary_cookie": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"strip_cookies": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_chunked": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cache_stale": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"permacache_storagezone": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"originshield_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"originshield_concurrency_limit": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"originshield_concurrency_requests": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(200),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"originshield_queue_requests": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(5000),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"originshield_queue_wait": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(30),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(3, 5, 15, 30, 45),
				},
			},
			"request_coalescing_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"request_coalescing_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(30),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(3, 5, 10, 15, 30, 60),
				},
			},
			"cors_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"cors_extensions": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(pullzoneCorsExtensionsDefault),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"limit_download_speed": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(0.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"limit_requests": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"limit_after": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(0.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"limit_burst": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"limit_connections": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"limit_bandwidth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
				Default:  booldefault.StaticBool(true),
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
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_minify_js": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"optimizer_smartimage": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
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
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"safehop_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"safehop_retry_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 2),
				},
			},
			"safehop_retry_delay": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(0, 1, 3, 5, 10),
				},
			},
			"safehop_retry_reasons": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(pullzoneSafehopRetryReasonsDefault),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"safehop_connection_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(10),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"safehop_response_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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

	// caching
	{
		// CacheVary
		vary := dataTf.CacheVary.Elements()
		dataApi.IgnoreQueryStrings = true
		for _, v := range vary {
			if v.(types.String).ValueString() == "querystring" {
				dataApi.IgnoreQueryStrings = false
			}
			if v.(types.String).ValueString() == "webp" {
				dataApi.EnableWebPVary = true
			}
			if v.(types.String).ValueString() == "country" {
				dataApi.EnableCountryCodeVary = true
			}
			if v.(types.String).ValueString() == "hostname" {
				dataApi.EnableHostnameVary = true
			}
			if v.(types.String).ValueString() == "mobile" {
				dataApi.EnableMobileVary = true
			}
			if v.(types.String).ValueString() == "avif" {
				dataApi.EnableAvifVary = true
			}
			if v.(types.String).ValueString() == "cookie" {
				dataApi.EnableCookieVary = true
			}
		}

		// CacheVaryQueryStringValues
		varyQueryString := []string{}
		for _, v := range dataTf.CacheVaryQueryStringValues.Elements() {
			varyQueryString = append(varyQueryString, v.(types.String).ValueString())
		}

		// CacheVaryCookieValues
		varyCookie := []string{}
		for _, v := range dataTf.CacheVaryCookieValues.Elements() {
			varyCookie = append(varyCookie, v.(types.String).ValueString())
		}

		// CacheStale
		stale := dataTf.CacheStale.Elements()
		for _, v := range stale {
			if v.(types.String).ValueString() == "offline" {
				dataApi.UseStaleWhileOffline = true
			}
			if v.(types.String).ValueString() == "updating" {
				dataApi.UseStaleWhileUpdating = true
			}
		}

		dataApi.EnableSmartCache = dataTf.CacheEnabled.ValueBool()
		dataApi.CacheControlMaxAgeOverride = dataTf.CacheExpirationTime.ValueInt64()
		dataApi.CacheControlPublicMaxAgeOverride = dataTf.CacheExpirationTimeBrowser.ValueInt64()
		dataApi.EnableQueryStringOrdering = dataTf.SortQueryString.ValueBool()
		dataApi.CacheErrorResponses = dataTf.CacheErrors.ValueBool()
		dataApi.QueryStringVaryParameters = varyQueryString
		dataApi.CookieVaryParameters = varyCookie
		dataApi.DisableCookies = dataTf.StripCookies.ValueBool()
		dataApi.EnableCacheSlice = dataTf.CacheChunked.ValueBool()
		dataApi.PermaCacheStorageZoneId = uint64(dataTf.PermacacheStoragezone.ValueInt64())
		dataApi.EnableOriginShield = dataTf.OriginShieldEnabled.ValueBool()
		dataApi.OriginShieldEnableConcurrencyLimit = dataTf.OriginShieldConcurrencyLimit.ValueBool()
		dataApi.OriginShieldMaxConcurrentRequests = uint64(dataTf.OriginShieldConcurrencyRequests.ValueInt64())
		dataApi.OriginShieldMaxQueuedRequests = uint64(dataTf.OriginShieldQueueRequests.ValueInt64())
		dataApi.OriginShieldQueueMaxWaitTime = uint64(dataTf.OriginShieldQueueWait.ValueInt64())
		dataApi.EnableRequestCoalescing = dataTf.RequestCoalescingEnabled.ValueBool()
		dataApi.RequestCoalescingTimeout = uint64(dataTf.RequestCoalescingTimeout.ValueInt64())
	}

	// cors
	{
		values := []string{}
		for _, extension := range dataTf.CorsExtensions.Elements() {
			values = append(values, extension.(types.String).ValueString())
		}
		slices.Sort(values)

		dataApi.EnableAccessControlOriginHeader = dataTf.CorsEnabled.ValueBool()
		dataApi.AccessControlOriginHeaderExtensions = values
	}

	// limits
	dataApi.LimitRatePerSecond = dataTf.LimitDownloadSpeed.ValueFloat64()
	dataApi.RequestLimit = uint64(dataTf.LimitRequests.ValueInt64())
	dataApi.LimitRateAfter = dataTf.LimitAfter.ValueFloat64()
	dataApi.BurstSize = uint64(dataTf.LimitBurst.ValueInt64())
	dataApi.ConnectionLimitPerIPCount = uint64(dataTf.LimitConnections.ValueInt64())
	dataApi.MonthlyBandwidthLimit = uint64(dataTf.LimitBandwidth.ValueInt64())

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

	// safe hop
	{
		dataApi.EnableSafeHop = dataTf.SafehopEnabled.ValueBool()
		dataApi.OriginRetries = uint8(dataTf.SafehopRetryCount.ValueInt64())
		dataApi.OriginRetryDelay = uint64(dataTf.SafehopRetryDelay.ValueInt64())
		dataApi.OriginConnectTimeout = uint64(dataTf.SafehopConnectionTimeout.ValueInt64())
		dataApi.OriginResponseTimeout = uint64(dataTf.SafehopResponseTimeout.ValueInt64())

		reasons := dataTf.SafehopRetryReasons.Elements()
		for _, reason := range reasons {
			if reason.(types.String).ValueString() == "connectionTimeout" {
				dataApi.OriginRetryConnectionTimeout = true
			}
			if reason.(types.String).ValueString() == "5xxResponse" {
				dataApi.OriginRetry5XXResponses = true
			}
			if reason.(types.String).ValueString() == "responseTimeout" {
				dataApi.OriginRetryResponseTimeout = true
			}
		}
	}

	return dataApi
}

func (r *PullzoneResource) convertApiToModel(dataApi api.Pullzone) (PullzoneResourceModel, diag.Diagnostics) {
	dataTf := PullzoneResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)

	// caching
	{
		// CacheVary
		var varyValues []attr.Value
		if !dataApi.IgnoreQueryStrings {
			varyValues = append(varyValues, types.StringValue("querystring"))
		}

		if dataApi.EnableWebPVary {
			varyValues = append(varyValues, types.StringValue("webp"))
		}

		if dataApi.EnableCountryCodeVary {
			varyValues = append(varyValues, types.StringValue("country"))
		}

		if dataApi.EnableHostnameVary {
			varyValues = append(varyValues, types.StringValue("hostname"))
		}

		if dataApi.EnableMobileVary {
			varyValues = append(varyValues, types.StringValue("mobile"))
		}

		if dataApi.EnableAvifVary {
			varyValues = append(varyValues, types.StringValue("avif"))
		}

		if dataApi.EnableCookieVary {
			varyValues = append(varyValues, types.StringValue("cookie"))
		}

		vary, diags := types.SetValue(types.StringType, varyValues)
		if diags != nil {
			return dataTf, diags
		}

		// CacheVaryQueryStringParams
		var varyQueryStringValues []attr.Value
		for _, qsParam := range dataApi.QueryStringVaryParameters {
			varyQueryStringValues = append(varyQueryStringValues, types.StringValue(qsParam))
		}

		varyQueryString, diags := types.SetValue(types.StringType, varyQueryStringValues)
		if diags != nil {
			return dataTf, diags
		}

		// CacheVaryCookieParams
		var varyCookieValues []attr.Value
		for _, cookieParam := range dataApi.CookieVaryParameters {
			varyCookieValues = append(varyCookieValues, types.StringValue(cookieParam))
		}

		varyCookie, diags := types.SetValue(types.StringType, varyCookieValues)
		if diags != nil {
			return dataTf, diags
		}

		// CacheStale
		var staleValues []attr.Value
		if dataApi.UseStaleWhileOffline {
			staleValues = append(staleValues, types.StringValue("offline"))
		}

		if dataApi.UseStaleWhileUpdating {
			staleValues = append(staleValues, types.StringValue("updating"))
		}

		stale, diags := types.SetValue(types.StringType, staleValues)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.CacheEnabled = types.BoolValue(dataApi.EnableSmartCache)
		dataTf.CacheExpirationTime = types.Int64Value(dataApi.CacheControlMaxAgeOverride)
		dataTf.CacheExpirationTimeBrowser = types.Int64Value(dataApi.CacheControlPublicMaxAgeOverride)
		dataTf.SortQueryString = types.BoolValue(dataApi.EnableQueryStringOrdering)
		dataTf.CacheErrors = types.BoolValue(dataApi.CacheErrorResponses)
		dataTf.CacheVary = vary
		dataTf.CacheVaryQueryStringValues = varyQueryString
		dataTf.CacheVaryCookieValues = varyCookie
		dataTf.StripCookies = types.BoolValue(dataApi.DisableCookies)
		dataTf.CacheChunked = types.BoolValue(dataApi.EnableCacheSlice)
		dataTf.CacheStale = stale
		dataTf.PermacacheStoragezone = types.Int64Value(int64(dataApi.PermaCacheStorageZoneId))
		dataTf.OriginShieldEnabled = types.BoolValue(dataApi.EnableOriginShield)
		dataTf.OriginShieldConcurrencyLimit = types.BoolValue(dataApi.OriginShieldEnableConcurrencyLimit)
		dataTf.OriginShieldConcurrencyRequests = types.Int64Value(int64(dataApi.OriginShieldMaxConcurrentRequests))
		dataTf.OriginShieldQueueRequests = types.Int64Value(int64(dataApi.OriginShieldMaxQueuedRequests))
		dataTf.OriginShieldQueueWait = types.Int64Value(int64(dataApi.OriginShieldQueueMaxWaitTime))
		dataTf.RequestCoalescingEnabled = types.BoolValue(dataApi.EnableRequestCoalescing)
		dataTf.RequestCoalescingTimeout = types.Int64Value(int64(dataApi.RequestCoalescingTimeout))
	}

	// cors
	{
		var extensionValues []attr.Value
		for _, extension := range dataApi.AccessControlOriginHeaderExtensions {
			extensionValues = append(extensionValues, types.StringValue(extension))
		}

		extensions, diags := types.SetValue(types.StringType, extensionValues)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.CorsEnabled = types.BoolValue(dataApi.EnableAccessControlOriginHeader)
		dataTf.CorsExtensions = extensions
	}

	// limits
	dataTf.LimitDownloadSpeed = types.Float64Value(dataApi.LimitRatePerSecond)
	dataTf.LimitRequests = types.Int64Value(int64(dataApi.RequestLimit))
	dataTf.LimitAfter = types.Float64Value(dataApi.LimitRateAfter)
	dataTf.LimitBurst = types.Int64Value(int64(dataApi.BurstSize))
	dataTf.LimitConnections = types.Int64Value(int64(dataApi.ConnectionLimitPerIPCount))
	dataTf.LimitBandwidth = types.Int64Value(int64(dataApi.MonthlyBandwidthLimit))

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

	// safe hop
	{
		dataTf.SafehopEnabled = types.BoolValue(dataApi.EnableSafeHop)
		dataTf.SafehopRetryCount = types.Int64Value(int64(dataApi.OriginRetries))
		dataTf.SafehopRetryDelay = types.Int64Value(int64(dataApi.OriginRetryDelay))
		dataTf.SafehopConnectionTimeout = types.Int64Value(int64(dataApi.OriginConnectTimeout))
		dataTf.SafehopResponseTimeout = types.Int64Value(int64(dataApi.OriginResponseTimeout))

		var reasonsValues []attr.Value
		if dataApi.OriginRetryConnectionTimeout {
			reasonsValues = append(reasonsValues, types.StringValue("connectionTimeout"))
		}

		if dataApi.OriginRetry5XXResponses {
			reasonsValues = append(reasonsValues, types.StringValue("5xxResponse"))
		}

		if dataApi.OriginRetryResponseTimeout {
			reasonsValues = append(reasonsValues, types.StringValue("responseTimeout"))
		}

		reasons, diags := types.SetValue(types.StringType, reasonsValues)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.SafehopRetryReasons = reasons
	}

	return dataTf, nil
}
