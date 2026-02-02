// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/pullzoneshieldresourcevalidator"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"regexp"
	"strconv"
)

var _ resource.Resource = &PullzoneShieldResource{}
var _ resource.ResourceWithImportState = &PullzoneShieldResource{}
var _ resource.ResourceWithConfigValidators = &PullzoneShieldResource{}

func NewPullzoneShield() resource.Resource {
	return &PullzoneShieldResource{}
}

type PullzoneShieldResource struct {
	client *api.Client
}

type PullzoneShieldResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	PullzoneId   types.Int64  `tfsdk:"pullzone"`
	Tier         types.String `tfsdk:"tier"`
	Whitelabel   types.Bool   `tfsdk:"whitelabel"`
	AccessList   types.Set    `tfsdk:"access_list"`
	BotDetection types.Object `tfsdk:"bot_detection"`
	DDoS         types.Object `tfsdk:"ddos"`
	WAF          types.Object `tfsdk:"waf"`
}

var pullzoneShieldAccessListType = map[string]attr.Type{
	"id":     types.Int64Type,
	"action": types.StringType,
}

var pullzoneShieldDdosType = map[string]attr.Type{
	"level":            types.StringType,
	"mode":             types.StringType,
	"challenge_window": types.Int64Type,
}

var pullzoneShieldBotDetectionType = map[string]attr.Type{
	"mode":                    types.StringType,
	"fingerprint_sensitivity": types.Int64Type,
	"fingerprint_aggression":  types.Int64Type,
	"ip_sensitivity":          types.Int64Type,
	"request_integrity":       types.Int64Type,
	"complex_fingerprinting":  types.BoolType,
}

var pullzoneShieldDdosLevelMap = map[uint8]string{
	0: "Asleep",
	1: "Low",
	2: "Medium",
	3: "High",
	4: "Extreme",
}

var pullzoneShieldDdosModeMap = map[uint8]string{
	0: "Log",
	1: "Block",
}

var pullzoneShieldBotDetectionModeMap = map[uint8]string{
	0: "Log",
	1: "Challenge",
}

var pullzoneShieldWafType = map[string]attr.Type{
	"enabled":                       types.BoolType,
	"mode":                          types.StringType,
	"realtime_threat_intelligence":  types.BoolType,
	"log_headers":                   types.BoolType,
	"log_headers_excluded":          types.SetType{ElemType: types.StringType},
	"allowed_http_versions":         types.SetType{ElemType: types.StringType},
	"allowed_http_methods":          types.SetType{ElemType: types.StringType},
	"allowed_request_content_types": types.SetType{ElemType: types.StringType},
	"detection_sensitivity":         types.Int64Type,
	"execution_sensitivity":         types.Int64Type,
	"blocking_sensitivity":          types.Int64Type,
	"body_limit_request":            types.StringType,
	"body_limit_response":           types.StringType,
	"rules_disabled":                types.SetType{ElemType: types.StringType},
	"rules_logonly":                 types.SetType{ElemType: types.StringType},
}

var pullzoneShieldWafModeMap = map[uint8]string{
	0: "Log",
	1: "Block",
}

var pullzoneShieldWafLogHeadersExcludedDefault = utils.ConvertStringSliceToSetMust([]string{
	"Cookie",
	"Authorization",
	"Signature",
	"Credential",
	"AccessKey",
})

var pullzoneShieldWafAllowedHttpVersionsDefault = utils.ConvertStringSliceToSetMust(pullzoneShieldWafAllowedHttpVersionsOptions)
var pullzoneShieldWafAllowedHttpVersionsOptions = []string{
	"HTTP/1.0",
	"HTTP/1.1",
	"HTTP/2",
	"HTTP/2.0",
}

var pullzoneShieldWafAllowedHttpMethodsDefault = utils.ConvertStringSliceToSetMust(pullzoneShieldWafAllowedHttpMethodsOptions)
var pullzoneShieldWafAllowedHttpMethodsOptions = []string{
	"GET",
	"HEAD",
	"POST",
	"PUT",
	"PATCH",
	"OPTIONS",
	"DELETE",
	"CONNECT",
	"TRACE",
}

var pullzoneShieldWafAllowedRequestContentTypesDefault = utils.ConvertStringSliceToSetMust([]string{
	"application/x-www-form-urlencoded",
	"multipart/form-data",
	"multipart/related",
	"text/xml",
	"application/xml",
	"application/soap+xml",
	"application/x-amf",
	"application/json",
	"application/octet-stream",
	"application/csp-report",
	"application/xss-auditor-report",
	"text/plain",
})

// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WAFPayloadLimitAction")'
var pullzoneShieldWafBodyLimitMap = map[uint8]string{
	0: "Block",
	1: "Log",
	2: "Ignore",
}

func (r *PullzoneShieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_shield"
}

func (r *PullzoneShieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages Bunny Shield for a bunny.net pullzone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the Bunny Shield.",
			},
			"pullzone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The ID of the linked pullzone.",
			},
			"tier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(pullzoneshieldresourcevalidator.PlanTypeMap[0]),
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneshieldresourcevalidator.PlanTypeMap)...),
				},
				Description: generateMarkdownMapOptions(pullzoneshieldresourcevalidator.PlanTypeMap),
			},
			"whitelabel": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Replace our bunny.net branded block and challenge pages with a white-labelled experience.",
			},
		},
		Blocks: map[string]schema.Block{
			"access_list": schema.SetNestedBlock{
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					PlanModifiers: []planmodifier.Object{
						objectplanmodifier.UseStateForUnknown(),
					},
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Required:    true,
							Description: "The ID of the Access List.",
						},
						"action": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(maps.Values(pullzoneAccessListActionMap)...),
							},
							Description: generateMarkdownMapOptions(pullzoneAccessListActionMap),
						},
					},
				},
				Validators: []validator.Set{
					pullzoneshieldresourcevalidator.AccessListSetUniqueValues(),
				},
			},
			"bot_detection": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(pullzoneShieldBotDetectionModeMap[0]),
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneShieldBotDetectionModeMap)...),
						},
						Description: "Indicates the mode the Bot Detection engine is running. " + generateMarkdownMapOptions(pullzoneShieldBotDetectionModeMap),
					},
					"fingerprint_sensitivity": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(1),
						Description: "Adjusts how precisely browsers are checked for signs of automation.",
						Validators: []validator.Int64{
							int64validator.Between(0, 3),
						},
					},
					"fingerprint_aggression": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(1),
						Description: "Controls how assertively unusual fingerprints are treated as bots.",
						Validators: []validator.Int64{
							int64validator.Between(1, 3),
						},
					},
					"ip_sensitivity": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(1),
						Description: "Monitors IP behaviour, reputation, and rate patterns.",
						Validators: []validator.Int64{
							int64validator.Between(0, 3),
						},
					},
					"request_integrity": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(1),
						Description: "Analyzes request headers, query structure, and protocol anomalies.",
						Validators: []validator.Int64{
							int64validator.Between(0, 3),
						},
					},
					"complex_fingerprinting": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Combines advanced entropy analysis and cross-session consistency.",
					},
				},
				Description: "Configures Bot Detection settings.",
			},
			"ddos": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"level": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneShieldDdosLevelMap)...),
						},
						Description: generateMarkdownMapOptions(pullzoneShieldDdosLevelMap),
					},
					"mode": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(pullzoneShieldDdosModeMap[0]),
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneShieldDdosModeMap)...),
						},
						Description: "Indicates the mode the engine is running. " + generateMarkdownMapOptions(pullzoneShieldDdosModeMap),
					},
					"challenge_window": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(3600),
						Validators: []validator.Int64{
							int64validator.OneOf(15*60, 30*60, 60*60, 6*60*60, 12*60*60, 24*60*60),
						},
						Description: "The window of time a visitor can access your website after passing a challenge. Once the timeout expires, they'll face a new challenge.",
					},
				},
				Description: "Configures DDoS settings.",
			},
			"waf": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Required:    true,
						Description: "Indicates whether the WAF (Web Application Firewall) is enabled.",
					},
					"mode": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(pullzoneShieldWafModeMap[0]),
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneShieldWafModeMap)...),
						},
						Description: "Indicates the mode the engine is running. " + generateMarkdownMapOptions(pullzoneShieldWafModeMap),
					},
					"realtime_threat_intelligence": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Real-time Threat Intelligence delivers zero-day protection by instantly detecting and blocking emerging threats.",
					},
					"log_headers": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "When enabled, detected WAF audit logs will contain the full list of request headers sent during the request.",
					},
					"log_headers_excluded": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(pullzoneShieldWafLogHeadersExcludedDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Description: "The list of headers excluded from the logs. They will still be used for processing WAF rules.",
					},
					"allowed_http_versions": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(pullzoneShieldWafAllowedHttpVersionsDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(pullzoneShieldWafAllowedHttpVersionsOptions...),
							),
						},
						Description: "Indicates allowed HTTP versions.",
					},
					"allowed_http_methods": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(pullzoneShieldWafAllowedHttpMethodsDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.LengthAtLeast(1),
								stringvalidator.OneOf(pullzoneShieldWafAllowedHttpMethodsOptions...),
							),
						},
						Description: "Indicates allowed HTTP methods.",
					},
					"allowed_request_content_types": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(pullzoneShieldWafAllowedRequestContentTypesDefault),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.LengthAtLeast(1),
								stringvalidator.RegexMatches(regexp.MustCompile(`^([a-zA-Z0-9.+_-]+)/([a-zA-Z0-9.+_-]+)$`), "is not allowed. Expected format: ([a-zA-Z0-9.+_-]+)/([a-zA-Z0-9.+_-]+)"),
							),
						},
						Description: "Indicates allowed values for request Content-Type.",
					},
					"rules_disabled": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Description: "List of disabled WAF rules.",
					},
					"rules_logonly": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Description: "List of WAF rules that will not be blocked, but will be logged when triggered.",
					},
					"detection_sensitivity": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(1),
						Validators: []validator.Int64{
							int64validator.Between(1, 4),
						},
						Description: "Determines which severity level of rules will trigger a detection log.",
					},
					"execution_sensitivity": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(1),
						Validators: []validator.Int64{
							int64validator.Between(1, 4),
						},
						Description: "Determines which severity level of rules will trigger the rules and their action.",
					},
					"blocking_sensitivity": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(1),
						Validators: []validator.Int64{
							int64validator.Between(1, 4),
						},
						Description: "Determines which severity level of rules will block requests.",
					},
					"body_limit_request": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("Log"),
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneShieldWafBodyLimitMap)...),
						},
						Description: "Determines the action to take when the request body length exceeds your plan limit. " + generateMarkdownMapOptions(pullzoneShieldWafBodyLimitMap),
					},
					"body_limit_response": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("Ignore"),
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneShieldWafBodyLimitMap)...),
						},
						Description: "Determines the action to take when the response body length exceeds your plan limit. " + generateMarkdownMapOptions(pullzoneShieldWafBodyLimitMap),
					},
				},
				Description: "Configures WAF settings.",
			},
		},
	}
}

func (r *PullzoneShieldResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		pullzoneshieldresourcevalidator.BotDetection(),
		pullzoneshieldresourcevalidator.RealtimeThreatIntelligence(),
		pullzoneshieldresourcevalidator.Whitelabel(),
	}
}

func (r *PullzoneShieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneShieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneShieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneShield(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create pullzone shield", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created shield for pullzone %d", dataTf.PullzoneId.ValueInt64()))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneShieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneShieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetPullzoneShield(ctx, data.Id.ValueInt64())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching pullzone shield", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneShieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneShieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApiResult, err := r.client.UpdatePullzoneShield(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating pullzone shield", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApiResult)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneShieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneShieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePullzoneShield(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting pullzone shield", err.Error()))
	}
}

func (r *PullzoneShieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pullzoneId, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error converting ID to integer", err.Error()))
		return
	}

	id, err := r.client.GetPullzoneShieldIdByPullzone(pullzoneId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Could not find pullzone shield", err.Error()))
		return
	}

	dataApi, err := r.client.GetPullzoneShield(ctx, id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching pullzone shield", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneShieldResource) convertModelToApi(ctx context.Context, dataTf PullzoneShieldResourceModel) api.PullzoneShield {
	dataApi := api.PullzoneShield{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.PullzoneId = dataTf.PullzoneId.ValueInt64()
	dataApi.PlanType = mapValueToKey(pullzoneshieldresourcevalidator.PlanTypeMap, dataTf.Tier.ValueString())
	dataApi.WhiteLabelResponsePages = dataTf.Whitelabel.ValueBool()

	// access_list
	{
		elements := dataTf.AccessList.Elements()
		accessLists := make([]api.PullzoneShieldAccessList, 0, len(elements))

		for _, element := range elements {
			listAttrs := element.(types.Object).Attributes()
			item := api.PullzoneShieldAccessList{
				IsEnabled: true,
			}

			if v, ok := listAttrs["id"]; ok {
				item.Id = v.(types.Int64).ValueInt64()
			}

			if v, ok := listAttrs["name"]; ok {
				item.Name = v.(types.String).ValueString()
			}

			if v, ok := listAttrs["action"]; ok {
				item.Action = mapValueToKey(pullzoneAccessListActionMap, v.(types.String).ValueString())
			}

			accessLists = append(accessLists, item)
		}

		dataApi.AccessLists = accessLists
	}

	// bot_detection
	{
		attrs := dataTf.BotDetection.Attributes()

		if len(attrs) == 0 {
			dataApi.BotDetectionMode = 0
			dataApi.BotDetectionFingerprintSensitivity = 0
			dataApi.BotDetectionFingerprintAggression = 0
			dataApi.BotDetectionIPSensitivity = 0
			dataApi.BotDetectionRequestIntegrity = 0
			dataApi.BotDetectionComplexFingerprinting = false
		} else {
			dataApi.BotDetectionMode = mapValueToKey(pullzoneShieldBotDetectionModeMap, attrs["mode"].(types.String).ValueString())

			if v, ok := attrs["fingerprint_sensitivity"]; ok {
				dataApi.BotDetectionFingerprintSensitivity = uint8(v.(types.Int64).ValueInt64())
			}

			if v, ok := attrs["fingerprint_aggression"]; ok {
				dataApi.BotDetectionFingerprintAggression = uint8(v.(types.Int64).ValueInt64())
			}

			if v, ok := attrs["ip_sensitivity"]; ok {
				dataApi.BotDetectionIPSensitivity = uint8(v.(types.Int64).ValueInt64())
			}

			if v, ok := attrs["request_integrity"]; ok {
				dataApi.BotDetectionRequestIntegrity = uint8(v.(types.Int64).ValueInt64())
			}

			if v, ok := attrs["complex_fingerprinting"]; ok {
				dataApi.BotDetectionComplexFingerprinting = v.(types.Bool).ValueBool()
			}
		}
	}

	// ddos
	{
		attrs := dataTf.DDoS.Attributes()

		dataApi.DDoSLevel = mapValueToKey(pullzoneShieldDdosLevelMap, attrs["level"].(types.String).ValueString())
		dataApi.DDoSMode = mapValueToKey(pullzoneShieldDdosModeMap, attrs["mode"].(types.String).ValueString())
		dataApi.DDosChallengeWindow = attrs["challenge_window"].(types.Int64).ValueInt64()
	}

	// waf
	{
		attrs := dataTf.WAF.Attributes()
		dataApi.WafEnabled = attrs["enabled"].(types.Bool).ValueBool()
		dataApi.WafMode = mapValueToKey(pullzoneShieldWafModeMap, attrs["mode"].(types.String).ValueString())
		dataApi.WafRealtimeThreatIntelligenceEnabled = attrs["realtime_threat_intelligence"].(types.Bool).ValueBool()
		dataApi.WafLogHeaders = attrs["log_headers"].(types.Bool).ValueBool()
		dataApi.WafRequestBodyLimitAction = mapValueToKey(pullzoneShieldWafBodyLimitMap, attrs["body_limit_request"].(types.String).ValueString())
		dataApi.WafResponseBodyLimitAction = mapValueToKey(pullzoneShieldWafBodyLimitMap, attrs["body_limit_response"].(types.String).ValueString())

		{
			var headers []string
			attrs["log_headers_excluded"].(types.Set).ElementsAs(ctx, &headers, false)
			dataApi.WafLogHeadersExcluded = headers
		}

		{
			elements := attrs["allowed_http_versions"].(types.Set).Elements()
			allowedHttpVersions := make([]string, len(elements))
			for i, item := range elements {
				allowedHttpVersions[i] = item.(types.String).ValueString()
			}
			dataApi.WafAllowedHttpVersions = allowedHttpVersions
		}

		{
			elements := attrs["allowed_http_methods"].(types.Set).Elements()
			allowedHttpMethods := make([]string, len(elements))
			for i, item := range elements {
				allowedHttpMethods[i] = item.(types.String).ValueString()
			}
			dataApi.WafAllowedHttpMethods = allowedHttpMethods
		}

		{
			elements := attrs["allowed_request_content_types"].(types.Set).Elements()
			contentTypes := make([]string, len(elements))
			for i, item := range elements {
				contentTypes[i] = item.(types.String).ValueString()
			}
			dataApi.WafAllowedRequestContentTypes = contentTypes
		}

		{
			dataApi.WafRuleSensitivityDetection = uint8(attrs["detection_sensitivity"].(types.Int64).ValueInt64())
			dataApi.WafRuleSensitivityExecution = uint8(attrs["execution_sensitivity"].(types.Int64).ValueInt64())
			dataApi.WafRuleSensitivityBlocking = uint8(attrs["blocking_sensitivity"].(types.Int64).ValueInt64())
		}

		{
			dataApi.WafRulesDisabled = utils.ConvertSetToStringSlice(attrs["rules_disabled"].(types.Set))
			dataApi.WafRulesLogonly = utils.ConvertSetToStringSlice(attrs["rules_logonly"].(types.Set))
		}
	}

	return dataApi
}

func (r *PullzoneShieldResource) convertApiToModel(dataApi api.PullzoneShield) (PullzoneShieldResourceModel, diag.Diagnostics) {
	dataTf := PullzoneShieldResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Tier = types.StringValue(mapKeyToValue(pullzoneshieldresourcevalidator.PlanTypeMap, dataApi.PlanType))
	dataTf.Whitelabel = types.BoolValue(dataApi.WhiteLabelResponsePages)

	// access_list
	{
		setValues := make([]attr.Value, 0, len(dataApi.AccessLists))
		for _, list := range dataApi.AccessLists {
			obj, diags := types.ObjectValue(pullzoneShieldAccessListType, map[string]attr.Value{
				"id":     types.Int64Value(list.Id),
				"action": types.StringValue(mapKeyToValue(pullzoneAccessListActionMap, list.Action)),
			})

			if diags != nil {
				return PullzoneShieldResourceModel{}, diags
			}

			setValues = append(setValues, obj)
		}

		accessListsSet, diags := types.SetValue(types.ObjectType{AttrTypes: pullzoneShieldAccessListType}, setValues)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		dataTf.AccessList = accessListsSet
	}

	// bot_detection
	{
		if dataApi.BotDetectionMode == 0 && dataApi.BotDetectionRequestIntegrity == 0 && dataApi.BotDetectionIPSensitivity == 0 && dataApi.BotDetectionFingerprintSensitivity == 0 {
			dataTf.BotDetection = types.ObjectNull(pullzoneShieldBotDetectionType)
		} else {
			values := map[string]attr.Value{
				"mode":                    types.StringValue(mapKeyToValue(pullzoneShieldBotDetectionModeMap, dataApi.BotDetectionMode)),
				"request_integrity":       types.Int64Value(int64(dataApi.BotDetectionRequestIntegrity)),
				"ip_sensitivity":          types.Int64Value(int64(dataApi.BotDetectionIPSensitivity)),
				"fingerprint_sensitivity": types.Int64Value(int64(dataApi.BotDetectionFingerprintSensitivity)),
				"fingerprint_aggression":  types.Int64Value(int64(dataApi.BotDetectionFingerprintAggression)),
				"complex_fingerprinting":  types.BoolValue(dataApi.BotDetectionComplexFingerprinting),
			}

			obj, diags := types.ObjectValue(pullzoneShieldBotDetectionType, values)
			if diags != nil {
				return PullzoneShieldResourceModel{}, diags
			}

			dataTf.BotDetection = obj
		}
	}

	// ddos
	{
		values := map[string]attr.Value{
			"level":            types.StringValue(mapKeyToValue(pullzoneShieldDdosLevelMap, dataApi.DDoSLevel)),
			"mode":             types.StringValue(mapKeyToValue(pullzoneShieldDdosModeMap, dataApi.DDoSMode)),
			"challenge_window": types.Int64Value(dataApi.DDosChallengeWindow),
		}

		obj, diags := types.ObjectValue(pullzoneShieldDdosType, values)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		dataTf.DDoS = obj
	}

	// waf
	{
		// log_headers_excluded
		headerValues := make([]attr.Value, 0, len(dataApi.WafLogHeadersExcluded))
		for _, header := range dataApi.WafLogHeadersExcluded {
			headerValues = append(headerValues, types.StringValue(header))
		}

		excludedSet, diags := types.SetValue(types.StringType, headerValues)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		// allowed_http_versions
		allowedHttpVersionsValues := make([]attr.Value, 0, len(dataApi.WafAllowedHttpVersions))
		for _, item := range dataApi.WafAllowedHttpVersions {
			allowedHttpVersionsValues = append(allowedHttpVersionsValues, types.StringValue(item))
		}

		allowedHttpVersionsSet, diags := types.SetValue(types.StringType, allowedHttpVersionsValues)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		// allowed_http_methods
		allowedHttpMethodsValues := make([]attr.Value, 0, len(dataApi.WafAllowedHttpMethods))
		for _, item := range dataApi.WafAllowedHttpMethods {
			allowedHttpMethodsValues = append(allowedHttpMethodsValues, types.StringValue(item))
		}

		allowedHttpMethodsSet, diags := types.SetValue(types.StringType, allowedHttpMethodsValues)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		// allowed_request_content_types
		allowedRequestContentTypesValues := make([]attr.Value, 0, len(dataApi.WafAllowedRequestContentTypes))
		for _, item := range dataApi.WafAllowedRequestContentTypes {
			allowedRequestContentTypesValues = append(allowedRequestContentTypesValues, types.StringValue(item))
		}

		allowedRequestContentTypesSet, diags := types.SetValue(types.StringType, allowedRequestContentTypesValues)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		// rules_disabled
		rulesDisabled, diags := utils.ConvertStringSliceToSet(dataApi.WafRulesDisabled)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		// rules_logonly
		rulesLogonly, diags := utils.ConvertStringSliceToSet(dataApi.WafRulesLogonly)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		values := map[string]attr.Value{
			"enabled":                       types.BoolValue(dataApi.WafEnabled),
			"mode":                          types.StringValue(mapKeyToValue(pullzoneShieldWafModeMap, dataApi.WafMode)),
			"realtime_threat_intelligence":  types.BoolValue(dataApi.WafRealtimeThreatIntelligenceEnabled),
			"log_headers":                   types.BoolValue(dataApi.WafLogHeaders),
			"log_headers_excluded":          excludedSet,
			"allowed_http_versions":         allowedHttpVersionsSet,
			"allowed_http_methods":          allowedHttpMethodsSet,
			"allowed_request_content_types": allowedRequestContentTypesSet,
			"rules_disabled":                rulesDisabled,
			"rules_logonly":                 rulesLogonly,
			"detection_sensitivity":         types.Int64Value(int64(dataApi.WafRuleSensitivityDetection)),
			"execution_sensitivity":         types.Int64Value(int64(dataApi.WafRuleSensitivityExecution)),
			"blocking_sensitivity":          types.Int64Value(int64(dataApi.WafRuleSensitivityBlocking)),
			"body_limit_request":            types.StringValue(mapKeyToValue(pullzoneShieldWafBodyLimitMap, dataApi.WafRequestBodyLimitAction)),
			"body_limit_response":           types.StringValue(mapKeyToValue(pullzoneShieldWafBodyLimitMap, dataApi.WafResponseBodyLimitAction)),
		}

		obj, diags := types.ObjectValue(pullzoneShieldWafType, values)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		dataTf.WAF = obj
	}

	return dataTf, nil
}
