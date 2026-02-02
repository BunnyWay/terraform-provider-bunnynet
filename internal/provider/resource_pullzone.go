// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/customtype"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/pullzoneresourcevalidator"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strconv"
)

var _ resource.Resource = &PullzoneResource{}
var _ resource.ResourceWithImportState = &PullzoneResource{}
var _ resource.ResourceWithModifyPlan = &PullzoneResource{}

func NewPullzoneResource() resource.Resource {
	return &PullzoneResource{}
}

type PullzoneResource struct {
	client *api.Client
}

type PullzoneResourceModel struct {
	Id                                 types.Int64   `tfsdk:"id"`
	Name                               types.String  `tfsdk:"name"`
	CdnDomain                          types.String  `tfsdk:"cdn_domain"`
	DisableLetsEncrypt                 types.Bool    `tfsdk:"disable_letsencrypt"`
	UseBackgroundUpdate                types.Bool    `tfsdk:"use_background_update"`
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
	OriginShieldZone                   types.String  `tfsdk:"originshield_zone"`
	RequestCoalescingEnabled           types.Bool    `tfsdk:"request_coalescing_enabled"`
	RequestCoalescingTimeout           types.Int64   `tfsdk:"request_coalescing_timeout"`
	CorsEnabled                        types.Bool    `tfsdk:"cors_enabled"`
	CorsExtensions                     types.Set     `tfsdk:"cors_extensions"`
	AddCanonicalHeader                 types.Bool    `tfsdk:"add_canonical_header"`
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
	OptimizerHtmlPrerender             types.Bool    `tfsdk:"optimizer_html_prerender"`
	OptimizerBurrow                    types.Bool    `tfsdk:"optimizer_burrow"`
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
	BlockRootPath                      types.Bool    `tfsdk:"block_root_path"`
	BlockPostRequests                  types.Bool    `tfsdk:"block_post_requests"`
	ReferersAllowed                    types.Set     `tfsdk:"allow_referers"`
	ReferersBlocked                    types.Set     `tfsdk:"block_referers"`
	BlockNoReferer                     types.Bool    `tfsdk:"block_no_referer"`
	IPsBlocked                         types.Set     `tfsdk:"block_ips"`
	LogEnabled                         types.Bool    `tfsdk:"log_enabled"`
	LogAnonymized                      types.Bool    `tfsdk:"log_anonymized"`
	LogAnonymizedStyle                 types.String  `tfsdk:"log_anonymized_style"`
	LogForwardEnabled                  types.Bool    `tfsdk:"log_forward_enabled"`
	LogForwardServer                   types.String  `tfsdk:"log_forward_server"`
	LogForwardPort                     types.Int64   `tfsdk:"log_forward_port"`
	LogForwardToken                    types.String  `tfsdk:"log_forward_token"`
	LogForwardProtocol                 types.String  `tfsdk:"log_forward_protocol"`
	LogForwardFormat                   types.String  `tfsdk:"log_forward_format"`
	LogStorageEnabled                  types.Bool    `tfsdk:"log_storage_enabled"`
	LogStorageZone                     types.Int64   `tfsdk:"log_storage_zone"`
	TlsSupport                         types.Set     `tfsdk:"tls_support"`
	ErrorPageWhitelabel                types.Bool    `tfsdk:"errorpage_whitelabel"`
	ErrorPageStatuspageEnabled         types.Bool    `tfsdk:"errorpage_statuspage_enabled"`
	ErrorPageStatuspageCode            types.String  `tfsdk:"errorpage_statuspage_code"`
	ErrorPageCustomEnabled             types.Bool    `tfsdk:"errorpage_custom_enabled"`
	ErrorPageCustomContent             types.String  `tfsdk:"errorpage_custom_content"`
	S3AuthEnabled                      types.Bool    `tfsdk:"s3_auth_enabled"`
	S3AuthKey                          types.String  `tfsdk:"s3_auth_key"`
	S3AuthSecret                       types.String  `tfsdk:"s3_auth_secret"`
	S3AuthRegion                       types.String  `tfsdk:"s3_auth_region"`
	TokenAuthEnabled                   types.Bool    `tfsdk:"token_auth_enabled"`
	TokenAuthIpValidation              types.Bool    `tfsdk:"token_auth_ip_validation"`
	TokenAuthKey                       types.String  `tfsdk:"token_auth_key"`
	WebsocketsEnabled                  types.Bool    `tfsdk:"websockets_enabled"`
	WebsocketsMaxConnections           types.Int64   `tfsdk:"websockets_max_connections"`
}

var pullzoneOriginTypes = map[string]attr.Type{
	"type":                  types.StringType,
	"url":                   customtype.PullzoneOriginUrlType{},
	"storagezone":           types.Int64Type,
	"follow_redirects":      types.BoolType,
	"host_header":           types.StringType,
	"forward_host_header":   types.BoolType,
	"verify_ssl":            types.BoolType,
	"script":                types.Int64Type,
	"middleware_script":     types.Int64Type,
	"container_app_id":      types.StringType,
	"container_endpoint_id": types.StringType,
}

var pullzoneRoutingTypes = map[string]attr.Type{
	"tier": types.StringType,
	"zones": types.SetType{
		ElemType: types.StringType,
	},
	"filters": types.SetType{
		ElemType: types.StringType,
	},
	"blocked_countries": types.SetType{
		ElemType: types.StringType,
	},
	"redirected_countries": types.SetType{
		ElemType: types.StringType,
	},
}

var pullzoneOriginTypeMap = map[uint8]string{
	0: "OriginUrl",
	1: "DnsAccelerate",
	2: "StorageZone",
	4: "ComputeScript",
	5: "ComputeContainer",
}

var pullzoneRoutingTierMap = map[uint8]string{
	0: "Standard",
	1: "Volume",
}

var pullzoneCacheVaryOptions = []string{"querystring", "webp", "country", "state", "hostname", "mobile", "avif", "cookie"}
var pullzoneCacheStaleOptions = []string{"offline", "updating"}
var pullzoneTlsSupportOptions = []string{"TLSv1.0", "TLSv1.1"}
var pullzoneSafehopRetryReasonsOptions = []string{"connectionTimeout", "5xxResponse", "responseTimeout"}
var pullzoneRoutingZonesOptions = []string{"AF", "ASIA", "EU", "SA", "US"}
var pullzoneRoutingFiltersOptions = []string{"all", "eu", "scripting"}
var pullzoneCorsExtensionsDefault = []string{"css", "eot", "gif", "jpeg", "jpg", "js", "mp3", "mp4", "mpeg", "png", "svg", "ttf", "webm", "webp", "woff", "woff2"}
var pullzoneOriginShieldZoneOptions = []string{"IL", "FR"}

var pullzoneLogAnonymizedStyleMap = map[uint8]string{
	0: "OneDigit",
	1: "Drop",
}

var pullzoneLogForwardProtocolMap = map[uint8]string{
	0: "UDP",
	1: "TCP",
	2: "TCPEncrypted",
	3: "DataDog",
}

var pullzoneLogForwardFormatMap = map[uint8]string{
	0: "Plain",
	1: "JSON",
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
	emptySetDefault := types.SetValueMust(types.StringType, []attr.Value{})

	pullzoneTlsSupportDefault, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue("TLSv1.0"),
		types.StringValue("TLSv1.1"),
	})

	if diags != nil {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}

	pullzoneCorsExtensionsSetDefault, diags := utils.ConvertStringSliceToSet(pullzoneCorsExtensionsDefault)
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
		Description: "This resource manages a bunny.net pullzone. Pullzones fetch content from the origin server and deliver it to end-users.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique ID of the pull zone.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The name of the pull zone.",
			},
			"cdn_domain": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The CNAME domain of the pull zone for setting up custom hostnames",
			},
			"disable_letsencrypt": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "If true, the built-in let's encrypt is disabled and requests are passed to the origin.",
			},
			"use_background_update": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether cache update is performed in the background.",
			},
			"cache_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether smart caching is enabled.",
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
				Description: "The override cache time, in seconds.",
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
				Description: "The override cache time for the end client, in seconds.",
			},
			"sort_querystring": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "If enabled, the query parameters will be automatically sorted into a consistent order before checking the cache.",
			},
			"cache_errors": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether bunny.net should be caching error responses.",
			},
			"cache_vary": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(pullzoneCacheVaryOptions...),
					),
				},
				MarkdownDescription: generateMarkdownSliceOptions(pullzoneCacheVaryOptions),
			},
			"cache_vary_querystring": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "Contains the list of vary parameters that will be used for vary cache by query string. If empty, all parameters will be used to construct the key",
			},
			"cache_vary_cookie": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				Description: "Contains the list of vary parameters that will be used for vary cache by cookie string. If empty, cookie vary will not be used.",
			},
			"strip_cookies": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "If enabled, bunny.net will strip all the Set-Cookie headers from the HTTP responses.",
			},
			"cache_chunked": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the cache slice (Optimize for large object delivery) feature is enabled for the Pull Zone",
			},
			"cache_stale": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(pullzoneCacheStaleOptions...),
					),
				},
				MarkdownDescription: generateMarkdownSliceOptions(pullzoneCacheStaleOptions),
			},
			"permacache_storagezone": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The storage zone ID for Perma-Cache.",
			},
			"originshield_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether Origin Shield is enabled.",
			},
			"originshield_concurrency_limit": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether there is a concurrency limit for Origin Shield.",
			},
			"originshield_concurrency_requests": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(200),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The number of concurrent requests for Origin Shield.",
			},
			"originshield_queue_requests": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(5000),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The number of queued requests for Origin Shield.",
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
				Description: "The maximum wait time for queued requests in Origin Shield, in seconds.",
			},
			"originshield_zone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(pullzoneOriginShieldZoneOptions...),
				},
				MarkdownDescription: generateMarkdownSliceOptions(pullzoneOriginShieldZoneOptions),
			},
			"request_coalescing_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether request coalescing is enabled.",
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
				Description: "Specifies the timeout period, in seconds, for request coalescing, determining how long to wait before sending combined requests to the origin.",
			},
			"block_root_path": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "This property indicates whether to block the root path.",
			},
			"block_post_requests": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether to block POST requests.",
			},
			"block_referers": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				Description: "The list of referrer hostnames that are blocked to access the pull zone. Requests containing the header \"Referer: hostname\" that is not on the list will be rejected. If empty, all the referrers are allowed.",
			},
			"allow_referers": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				Description: "The list of referrer hostnames that are allowed to access the pull zone. Requests containing the header \"Referer: hostname\" that is not on the list will be rejected. If empty, all the referrers are allowed.",
			},
			"block_no_referer": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether requests without a referer should be blocked.",
			},
			"block_ips": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				Description: "The list of IPs that are blocked from accessing the pull zone. Requests coming from the following IPs will be rejected. If empty, all the IPs will be allowed",
			},
			"log_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether logging is enabled.",
			},
			"log_anonymized": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether logs are anonymized.",
			},
			"log_anonymized_style": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("OneDigit"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneLogAnonymizedStyleMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneLogAnonymizedStyleMap),
			},
			"log_forward_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether log forwarding is enabled.",
			},
			"log_forward_server": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The server address for log forwarding.",
			},
			"log_forward_port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
				Description: "The port number for log forwarding.",
			},
			"log_forward_token": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The token used for log forwarding authentication.",
			},
			"log_forward_protocol": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("UDP"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneLogForwardProtocolMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneLogForwardProtocolMap),
			},
			"log_forward_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("JSON"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneLogForwardFormatMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneLogForwardFormatMap),
			},
			"log_storage_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether log storage is enabled.",
			},
			"log_storage_zone": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The storage zone ID for log storage.",
			},
			"tls_support": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(pullzoneTlsSupportDefault),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(pullzoneTlsSupportOptions...),
					),
				},
				MarkdownDescription: generateMarkdownSliceOptions(pullzoneTlsSupportOptions),
			},
			"errorpage_whitelabel": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the error pages should be white-labelled or not",
			},
			"errorpage_statuspage_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the statuspage widget should be displayed on the error pages.",
			},
			"errorpage_statuspage_code": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The statuspage code that will be used to build the status widget.",
			},
			"errorpage_custom_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether custom error page code should be enabled.",
			},
			"errorpage_custom_content": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Contains the custom error page code that will be returned.",
			},
			"s3_auth_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether requests to origin will be signed with AWS Signature Version 4.",
			},
			"s3_auth_key": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The access key used to authenticate the requests.",
			},
			"s3_auth_secret": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The secret key used to authenticate the requests.",
			},
			"s3_auth_region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The region name of the bucket used to authenticate the requests.",
			},
			"token_auth_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether requests without a valid token and expiry timestamp will be rejected.",
			},
			"token_auth_ip_validation": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the secure hash generated by the server will also include an IP address.",
			},
			"token_auth_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The auth key used for secure URL token authentication.",
			},
			"cors_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether CORS (Cross-Origin Resource Sharing) is enabled.",
			},
			"cors_extensions": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(pullzoneCorsExtensionsSetDefault),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				Description: "A list of file extensions for which CORS is enabled.",
			},
			"add_canonical_header": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the Canonical header is added to the responses.",
			},
			"limit_download_speed": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(0.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Float64{
					float64validator.AtLeast(0.0),
				},
				Description: "The maximum download speed, in kb/s. Use 0 for unlimited.",
			},
			"limit_requests": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The maximum amount of requests per IP per second.",
			},
			"limit_after": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(0.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Float64{
					float64validator.AtLeast(0.0),
				},
				Description: "The amount of data after the rate limit will be activated.",
			},
			"limit_burst": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "Excessive requests are delayed until their number exceeds the maximum burst size.",
			},
			"limit_connections": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The number of connections limited per IP.",
			},
			"limit_bandwidth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The maximum bandwidth limit in bytes.",
			},
			"optimizer_classes_force": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the optimizer class list should be enforced.",
			},
			"optimizer_dynamic_image_api": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the image manipulation should be enabled.",
			},
			"optimizer_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether Bunny Optimizer should be enabled.",
			},
			"optimizer_minify_css": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the CSS minification should be enabled.",
			},
			"optimizer_minify_js": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the JavaScript minification should be enabled.",
			},
			"optimizer_html_prerender": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether HTML Prerender should be enabled.",
			},
			"optimizer_burrow": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether Burrow Smart Routing should be enabled.",
			},
			"optimizer_smartimage": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the automatic image optimization should be enabled.",
			},
			"optimizer_smartimage_desktop_maxwidth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1600),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The maximum automatic image size for desktop clients.",
			},
			"optimizer_smartimage_desktop_quality": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(85),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
				Description: "The image quality for desktop clients.",
			},
			"optimizer_smartimage_mobile_maxwidth": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(800),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The maximum automatic image size for mobile clients.",
			},
			"optimizer_smartimage_mobile_quality": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(70),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
				Description: "Determines the image quality for mobile clients",
			},
			"optimizer_watermark": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether image watermarking should be enabled.",
			},
			"optimizer_watermark_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					// @TODO validate URL
				},
				Description: "The URL of the watermark image.",
			},
			"optimizer_watermark_position": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("BottomLeft"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneOptimizerWatermarkPositionMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneOptimizerWatermarkPositionMap),
			},
			"optimizer_watermark_borderoffset": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(3.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Description: "The offset of the watermark image.",
			},
			"optimizer_watermark_minsize": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(300),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The minimum image size to which the watermark will be added.",
			},
			"optimizer_webp": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the WebP optimization should be enabled.",
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
				Description: "The number of retries to the origin server.",
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
				Description: "The amount of time that the CDN should wait before retrying an origin request.",
			},
			"safehop_retry_reasons": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(pullzoneSafehopRetryReasonsDefault),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(pullzoneSafehopRetryReasonsOptions...),
					),
				},
				MarkdownDescription: generateMarkdownSliceOptions(pullzoneSafehopRetryReasonsOptions),
			},
			"safehop_connection_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(10),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(3, 5, 10),
				},
				Description: "The amount of seconds to wait when connecting to the origin. Otherwise the request will fail or retry.",
			},
			"safehop_response_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(5, 15, 30, 45, 60),
				},
				Description: "The amount of seconds to wait when waiting for the origin reply. Otherwise the request will fail or retry.",
			},
			"websockets_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the WebSocket support is enabled.",
			},
			"websockets_max_connections": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(500),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(500, 1000, 2500, 5000, 10000, 25000),
				},
				Description: "The maximum allowed concurrent WebSocket connections.",
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
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneOriginTypeMap)...),
						},
						MarkdownDescription: generateMarkdownMapOptions(pullzoneOriginTypeMap),
					},
					"url": schema.StringAttribute{
						CustomType:  customtype.PullzoneOriginUrlType{},
						Optional:    true,
						Description: "The origin URL from where the files are fetched.",
					},
					"storagezone": schema.Int64Attribute{
						Optional:    true,
						Description: "The ID of the linked storage zone.",
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"follow_redirects": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Indicates whether the zone will follow origin redirects.",
					},
					"host_header": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						Description: "The host header that will be sent to the origin.",
					},
					"forward_host_header": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Indicates whether the current hostname is forwarded to the origin.",
					},
					"verify_ssl": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Indicates whether the Origin's TLS certificate should be verified.",
					},
					"script": schema.Int64Attribute{
						Optional:    true,
						Description: "The ID of the linked compute script.",
					},
					"middleware_script": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "The ID of the compute script used as a middleware.",
					},
					"container_app_id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The ID if the compute container app.",
						Default:     stringdefault.StaticString(""),
					},
					"container_endpoint_id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						Description: "The ID if the compute container app endpoint.",
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
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneRoutingTierMap)...),
						},
						MarkdownDescription: generateMarkdownMapOptions(pullzoneRoutingTierMap),
					},
					"zones": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(pullzoneRoutingZonesDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(pullzoneRoutingZonesOptions...),
							),
						},
						MarkdownDescription: generateMarkdownSliceOptions(pullzoneRoutingZonesOptions),
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
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(pullzoneRoutingFiltersOptions...),
							),
						},
						MarkdownDescription: generateMarkdownSliceOptions(pullzoneRoutingFiltersOptions),
					},
					"blocked_countries": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(emptySetDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Description: "The list of blocked countries with the two-letter Alpha2 ISO codes. Traffic connecting from a blocked country will be rejected on the DNS level.",
					},
					"redirected_countries": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(emptySetDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Description: "The list of budget redirected countries with the two-letter Alpha2 ISO codes. Traffic from a redirected country will connect to the cheapest possible node in North America or Europe.",
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
			},
		},
	}
}

func (r *PullzoneResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		pullzoneresourcevalidator.OriginComputeScript(),
		pullzoneresourcevalidator.MiddlewareScript(),
		pullzoneresourcevalidator.PermacacheCacheExpirationTime(),
		pullzoneresourcevalidator.CacheStaleBackgroundUpdate(),
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
	pzMutex.Lock(0)
	dataApi, err := r.client.CreatePullzone(dataApi)
	pzMutex.Unlock(0)

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
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

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

func (r *PullzoneResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	// permacache_storagezone
	{
		var planPermacacheStoragezone types.Int64
		req.Plan.GetAttribute(ctx, path.Root("permacache_storagezone"), &planPermacacheStoragezone)

		if planPermacacheStoragezone.IsUnknown() {
			resp.Plan.SetAttribute(ctx, path.Root("cache_expiration_time"), types.Int64Unknown())
		} else if planPermacacheStoragezone.ValueInt64() > 0 {
			resp.Plan.SetAttribute(ctx, path.Root("cache_expiration_time"), pullzoneresourcevalidator.DefaultCacheExpirationTimeForPermacache)
		}
	}

	// cache_stale
	{
		var stateCacheStale []string
		req.State.GetAttribute(ctx, path.Root("cache_stale"), &stateCacheStale)

		var planCacheStale []string
		req.Plan.GetAttribute(ctx, path.Root("cache_stale"), &planCacheStale)

		if len(planCacheStale) > 0 {
			resp.Plan.SetAttribute(ctx, path.Root("use_background_update"), true)
		}

		if len(planCacheStale) == 0 && len(stateCacheStale) > 0 {
			resp.Plan.SetAttribute(ctx, path.Root("use_background_update"), false)
		}
	}
}

func (r *PullzoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.CorsExtensions.IsNull() {
		extensions, diags := utils.ConvertStringSliceToSet(pullzoneCorsExtensionsDefault)
		if diags != nil {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.CorsExtensions = extensions
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
	dataApi.DisableLetsEncrypt = dataTf.DisableLetsEncrypt.ValueBool()
	dataApi.UseBackgroundUpdate = dataTf.UseBackgroundUpdate.ValueBool()

	// caching
	{
		// CacheVary
		dataApi.IgnoreQueryStrings = true
		for _, v := range dataTf.CacheVary.Elements() {
			if v.(types.String).ValueString() == "querystring" {
				dataApi.IgnoreQueryStrings = false
			}
			if v.(types.String).ValueString() == "webp" {
				dataApi.EnableWebPVary = true
			}
			if v.(types.String).ValueString() == "country" {
				dataApi.EnableCountryCodeVary = true
			}
			if v.(types.String).ValueString() == "state" {
				dataApi.EnableCountryStateCodeVary = true
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

		// CacheStale
		for _, v := range dataTf.CacheStale.Elements() {
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
		dataApi.QueryStringVaryParameters = utils.ConvertSetToStringSlice(dataTf.CacheVaryQueryStringValues)
		dataApi.CookieVaryParameters = utils.ConvertSetToStringSlice(dataTf.CacheVaryCookieValues)
		dataApi.DisableCookies = dataTf.StripCookies.ValueBool()
		dataApi.EnableCacheSlice = dataTf.CacheChunked.ValueBool()
		dataApi.PermaCacheStorageZoneId = uint64(dataTf.PermacacheStoragezone.ValueInt64())
		dataApi.EnableOriginShield = dataTf.OriginShieldEnabled.ValueBool()
		dataApi.OriginShieldEnableConcurrencyLimit = dataTf.OriginShieldConcurrencyLimit.ValueBool()
		dataApi.OriginShieldMaxConcurrentRequests = uint64(dataTf.OriginShieldConcurrencyRequests.ValueInt64())
		dataApi.OriginShieldMaxQueuedRequests = uint64(dataTf.OriginShieldQueueRequests.ValueInt64())
		dataApi.OriginShieldQueueMaxWaitTime = uint64(dataTf.OriginShieldQueueWait.ValueInt64())
		dataApi.OriginShieldZoneCode = dataTf.OriginShieldZone.ValueString()
		dataApi.EnableRequestCoalescing = dataTf.RequestCoalescingEnabled.ValueBool()
		dataApi.RequestCoalescingTimeout = uint64(dataTf.RequestCoalescingTimeout.ValueInt64())
	}

	// security
	{
		dataApi.BlockRootPathAccess = dataTf.BlockRootPath.ValueBool()
		dataApi.BlockPostRequests = dataTf.BlockPostRequests.ValueBool()
		dataApi.AllowedReferrers = utils.ConvertSetToStringSlice(dataTf.ReferersAllowed)
		dataApi.BlockedReferrers = utils.ConvertSetToStringSlice(dataTf.ReferersBlocked)
		dataApi.BlockNoneReferrer = dataTf.BlockNoReferer.ValueBool()
		dataApi.BlockedIps = utils.ConvertSetToStringSlice(dataTf.IPsBlocked)
		dataApi.EnableLogging = dataTf.LogEnabled.ValueBool()
		dataApi.LoggingIPAnonymizationEnabled = dataTf.LogAnonymized.ValueBool()
		dataApi.LogAnonymizationType = mapValueToKey(pullzoneLogAnonymizedStyleMap, dataTf.LogAnonymizedStyle.ValueString())
		dataApi.LogForwardingEnabled = dataTf.LogForwardEnabled.ValueBool()
		dataApi.LogForwardingHostname = dataTf.LogForwardServer.ValueString()
		dataApi.LogForwardingPort = uint16(dataTf.LogForwardPort.ValueInt64())
		dataApi.LogForwardingToken = dataTf.LogForwardToken.ValueString()
		dataApi.LogForwardingProtocol = mapValueToKey(pullzoneLogForwardProtocolMap, dataTf.LogForwardProtocol.ValueString())
		dataApi.LogForwardingFormat = mapValueToKey(pullzoneLogForwardFormatMap, dataTf.LogForwardFormat.ValueString())
		dataApi.LoggingSaveToStorage = dataTf.LogStorageEnabled.ValueBool()
		dataApi.LoggingStorageZoneId = uint64(dataTf.LogStorageZone.ValueInt64())
		dataApi.ErrorPageWhitelabel = dataTf.ErrorPageWhitelabel.ValueBool()
		dataApi.ErrorPageEnableStatuspageWidget = dataTf.ErrorPageStatuspageEnabled.ValueBool()
		dataApi.ErrorPageStatuspageCode = dataTf.ErrorPageStatuspageCode.ValueString()
		dataApi.ErrorPageEnableCustomCode = dataTf.ErrorPageCustomEnabled.ValueBool()
		dataApi.ErrorPageCustomCode = dataTf.ErrorPageCustomContent.ValueString()
		dataApi.AWSSigningEnabled = dataTf.S3AuthEnabled.ValueBool()
		dataApi.AWSSigningKey = dataTf.S3AuthKey.ValueString()
		dataApi.AWSSigningSecret = dataTf.S3AuthSecret.ValueString()
		dataApi.AWSSigningRegionName = dataTf.S3AuthRegion.ValueString()
		dataApi.ZoneSecurityEnabled = dataTf.TokenAuthEnabled.ValueBool()
		dataApi.ZoneSecurityIncludeHashRemoteIP = dataTf.TokenAuthIpValidation.ValueBool()

		for _, v := range dataTf.TlsSupport.Elements() {
			if v.(types.String).ValueString() == "TLSv1.0" {
				dataApi.EnableTLS1 = true
			}
			if v.(types.String).ValueString() == "TLSv1.1" {
				dataApi.EnableTLS11 = true
			}
		}
	}

	// headers
	{
		extensions := utils.ConvertSetToStringSlice(dataTf.CorsExtensions)
		slices.Sort(extensions)

		dataApi.EnableAccessControlOriginHeader = dataTf.CorsEnabled.ValueBool()
		dataApi.AccessControlOriginHeaderExtensions = extensions
		dataApi.AddCanonicalHeader = dataTf.AddCanonicalHeader.ValueBool()
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
	dataApi.OptimizerPrerenderHtml = dataTf.OptimizerHtmlPrerender.ValueBool()
	dataApi.OptimizerTunnelEnabled = dataTf.OptimizerBurrow.ValueBool()
	dataApi.OptimizerEnableWebP = dataTf.OptimizerWebp.ValueBool()
	dataApi.OptimizerForceClasses = dataTf.OptimizerClassesForce.ValueBool()
	dataApi.OptimizerEnableManipulationEngine = dataTf.OptimizerDynamicImageApi.ValueBool()
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
	dataApi.OriginUrl = origin["url"].(customtype.PullzoneOriginUrlValue).ValueString()
	dataApi.StorageZoneId = origin["storagezone"].(types.Int64).ValueInt64()
	dataApi.OriginHostHeader = origin["host_header"].(types.String).ValueString()
	dataApi.AddHostHeader = origin["forward_host_header"].(types.Bool).ValueBool()
	dataApi.VerifyOriginSSL = origin["verify_ssl"].(types.Bool).ValueBool()
	dataApi.FollowRedirects = origin["follow_redirects"].(types.Bool).ValueBool()
	dataApi.EdgeScriptId = origin["script"].(types.Int64).ValueInt64()
	dataApi.MiddlewareScriptId = origin["middleware_script"].(types.Int64).ValueInt64()
	dataApi.MagicContainersAppId = origin["container_app_id"].(types.String).ValueString()
	dataApi.MagicContainersEndpointId = origin["container_endpoint_id"].(types.String).ValueString()

	// websockets
	dataApi.EnableWebSockets = dataTf.WebsocketsEnabled.ValueBool()
	dataApi.MaxWebSocketConnections = uint64(dataTf.WebsocketsMaxConnections.ValueInt64())

	// routing
	dataApi.Type = mapValueToKey(pullzoneRoutingTierMap, routing["tier"].(types.String).ValueString())
	dataApi.RoutingFilters = utils.ConvertSetToStringSlice(routing["filters"].(types.Set))
	dataApi.BlockedCountries = utils.ConvertSetToStringSlice(routing["blocked_countries"].(types.Set))
	dataApi.BudgetRedirectedCountries = utils.ConvertSetToStringSlice(routing["redirected_countries"].(types.Set))

	// routing.zones
	{
		for _, zone := range routing["zones"].(types.Set).Elements() {
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

	// safe hop
	{
		dataApi.EnableSafeHop = dataTf.SafehopEnabled.ValueBool()
		dataApi.OriginRetries = uint8(dataTf.SafehopRetryCount.ValueInt64())
		dataApi.OriginRetryDelay = uint64(dataTf.SafehopRetryDelay.ValueInt64())
		dataApi.OriginConnectTimeout = uint64(dataTf.SafehopConnectionTimeout.ValueInt64())
		dataApi.OriginResponseTimeout = uint64(dataTf.SafehopResponseTimeout.ValueInt64())

		for _, reason := range dataTf.SafehopRetryReasons.Elements() {
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
	return pullzoneApiToTf(dataApi)
}

func pullzoneApiToTf(dataApi api.Pullzone) (PullzoneResourceModel, diag.Diagnostics) {
	dataTf := PullzoneResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.CdnDomain = types.StringValue(dataApi.CnameDomain)
	dataTf.DisableLetsEncrypt = types.BoolValue(dataApi.DisableLetsEncrypt)
	dataTf.UseBackgroundUpdate = types.BoolValue(dataApi.UseBackgroundUpdate)

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

		if dataApi.EnableCountryStateCodeVary {
			varyValues = append(varyValues, types.StringValue("state"))
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
		varyQueryString, diags := utils.ConvertStringSliceToSet(dataApi.QueryStringVaryParameters)
		if diags != nil {
			return dataTf, diags
		}

		// CacheVaryCookieParams
		varyCookie, diags := utils.ConvertStringSliceToSet(dataApi.CookieVaryParameters)
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
		dataTf.OriginShieldZone = types.StringValue(dataApi.OriginShieldZoneCode)
		dataTf.RequestCoalescingEnabled = types.BoolValue(dataApi.EnableRequestCoalescing)
		dataTf.RequestCoalescingTimeout = types.Int64Value(int64(dataApi.RequestCoalescingTimeout))
	}

	// security
	{
		referersAllowed, diags := utils.ConvertStringSliceToSet(dataApi.AllowedReferrers)
		if diags != nil {
			return dataTf, diags
		}

		referersBlocked, diags := utils.ConvertStringSliceToSet(dataApi.BlockedReferrers)
		if diags != nil {
			return dataTf, diags
		}

		ipsBlocked, diags := utils.ConvertStringSliceToSet(dataApi.BlockedIps)
		if diags != nil {
			return dataTf, diags
		}

		// tls support
		{
			var tlsSupportValues []attr.Value
			if dataApi.EnableTLS1 {
				tlsSupportValues = append(tlsSupportValues, types.StringValue("TLSv1.0"))
			}
			if dataApi.EnableTLS11 {
				tlsSupportValues = append(tlsSupportValues, types.StringValue("TLSv1.1"))
			}

			tlsSupport, diags := types.SetValue(types.StringType, tlsSupportValues)
			if diags != nil {
				return dataTf, diags
			}

			dataTf.TlsSupport = tlsSupport
		}

		dataTf.BlockRootPath = types.BoolValue(dataApi.BlockRootPathAccess)
		dataTf.BlockPostRequests = types.BoolValue(dataApi.BlockPostRequests)
		dataTf.ReferersAllowed = referersAllowed
		dataTf.ReferersBlocked = referersBlocked
		dataTf.BlockNoReferer = types.BoolValue(dataApi.BlockNoneReferrer)
		dataTf.IPsBlocked = ipsBlocked
		dataTf.LogEnabled = types.BoolValue(dataApi.EnableLogging)
		dataTf.LogAnonymized = types.BoolValue(dataApi.LoggingIPAnonymizationEnabled)
		dataTf.LogAnonymizedStyle = types.StringValue(mapKeyToValue(pullzoneLogAnonymizedStyleMap, dataApi.LogAnonymizationType))
		dataTf.LogForwardEnabled = types.BoolValue(dataApi.LogForwardingEnabled)
		dataTf.LogForwardServer = types.StringValue(dataApi.LogForwardingHostname)
		dataTf.LogForwardPort = types.Int64Value(int64(dataApi.LogForwardingPort))
		dataTf.LogForwardToken = types.StringValue(dataApi.LogForwardingToken)
		dataTf.LogForwardProtocol = types.StringValue(mapKeyToValue(pullzoneLogForwardProtocolMap, dataApi.LogForwardingProtocol))
		dataTf.LogForwardFormat = types.StringValue(mapKeyToValue(pullzoneLogForwardFormatMap, dataApi.LogForwardingFormat))
		dataTf.LogStorageEnabled = types.BoolValue(dataApi.LoggingSaveToStorage)
		dataTf.LogStorageZone = types.Int64Value(int64(dataApi.LoggingStorageZoneId))

		dataTf.ErrorPageWhitelabel = types.BoolValue(dataApi.ErrorPageWhitelabel)
		dataTf.ErrorPageStatuspageEnabled = types.BoolValue(dataApi.ErrorPageEnableStatuspageWidget)
		dataTf.ErrorPageStatuspageCode = types.StringValue(dataApi.ErrorPageStatuspageCode)
		dataTf.ErrorPageCustomEnabled = types.BoolValue(dataApi.ErrorPageEnableCustomCode)
		dataTf.ErrorPageCustomContent = types.StringValue(dataApi.ErrorPageCustomCode)
		dataTf.S3AuthEnabled = types.BoolValue(dataApi.AWSSigningEnabled)
		dataTf.S3AuthKey = types.StringValue(dataApi.AWSSigningKey)
		dataTf.S3AuthSecret = types.StringValue(dataApi.AWSSigningSecret)
		dataTf.S3AuthRegion = types.StringValue(dataApi.AWSSigningRegionName)
		dataTf.TokenAuthEnabled = types.BoolValue(dataApi.ZoneSecurityEnabled)
		dataTf.TokenAuthIpValidation = types.BoolValue(dataApi.ZoneSecurityIncludeHashRemoteIP)
		dataTf.TokenAuthKey = types.StringValue(dataApi.ZoneSecurityKey)
	}

	// headers
	{
		extensions, diags := utils.ConvertStringSliceToSet(dataApi.AccessControlOriginHeaderExtensions)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.CorsEnabled = types.BoolValue(dataApi.EnableAccessControlOriginHeader)
		dataTf.CorsExtensions = extensions
		dataTf.AddCanonicalHeader = types.BoolValue(dataApi.AddCanonicalHeader)
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
	dataTf.OptimizerHtmlPrerender = types.BoolValue(dataApi.OptimizerPrerenderHtml)
	dataTf.OptimizerBurrow = types.BoolValue(dataApi.OptimizerTunnelEnabled)
	dataTf.OptimizerWebp = types.BoolValue(dataApi.OptimizerEnableWebP)
	dataTf.OptimizerClassesForce = types.BoolValue(dataApi.OptimizerForceClasses)
	dataTf.OptimizerDynamicImageApi = types.BoolValue(dataApi.OptimizerEnableManipulationEngine)
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
			"url":  customtype.PullzoneOriginUrlValue{StringValue: typeStringOrNull(dataApi.OriginUrl)},
		}

		if dataApi.StorageZoneId == 0 || dataApi.StorageZoneId == -1 {
			originValues["storagezone"] = types.Int64Null()
		} else {
			originValues["storagezone"] = types.Int64Value(dataApi.StorageZoneId)
		}

		if dataApi.EdgeScriptId <= 0 {
			originValues["script"] = types.Int64Null()
		} else {
			originValues["script"] = types.Int64Value(dataApi.EdgeScriptId)
		}

		originValues["middleware_script"] = types.Int64Value(dataApi.MiddlewareScriptId)
		originValues["follow_redirects"] = types.BoolValue(dataApi.FollowRedirects)
		originValues["host_header"] = types.StringValue(dataApi.OriginHostHeader)
		originValues["forward_host_header"] = types.BoolValue(dataApi.AddHostHeader)
		originValues["verify_ssl"] = types.BoolValue(dataApi.VerifyOriginSSL)
		originValues["container_app_id"] = types.StringValue(dataApi.MagicContainersAppId)
		originValues["container_endpoint_id"] = types.StringValue(dataApi.MagicContainersEndpointId)

		origin, diags := types.ObjectValue(pullzoneOriginTypes, originValues)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.Origin = origin
	}

	// websockets
	dataTf.WebsocketsEnabled = types.BoolValue(dataApi.EnableWebSockets)
	dataTf.WebsocketsMaxConnections = types.Int64Value(int64(dataApi.MaxWebSocketConnections))

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
			filters, diags := utils.ConvertStringSliceToSet(dataApi.RoutingFilters)
			if diags != nil {
				return dataTf, diags
			}
			routingValues["filters"] = filters
		}

		// blocked countries
		{
			blockedCountries, diags := utils.ConvertStringSliceToSet(dataApi.BlockedCountries)
			if diags != nil {
				return dataTf, diags
			}
			routingValues["blocked_countries"] = blockedCountries
		}

		// redirected countries
		{
			redirectedCountries, diags := utils.ConvertStringSliceToSet(dataApi.BudgetRedirectedCountries)
			if diags != nil {
				return dataTf, diags
			}
			routingValues["redirected_countries"] = redirectedCountries
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
