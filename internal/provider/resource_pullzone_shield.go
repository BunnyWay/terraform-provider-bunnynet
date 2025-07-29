// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"regexp"
	"strconv"
)

var _ resource.Resource = &PullzoneShieldResource{}
var _ resource.ResourceWithImportState = &PullzoneShieldResource{}

func NewPullzoneShield() resource.Resource {
	return &PullzoneShieldResource{}
}

type PullzoneShieldResource struct {
	client *api.Client
}

type PullzoneShieldResourceModel struct {
	Id         types.Int64  `tfsdk:"id"`
	PullzoneId types.Int64  `tfsdk:"pullzone"`
	Tier       types.String `tfsdk:"tier"`
	Whitelabel types.Bool   `tfsdk:"whitelabel"`
	DDoS       types.Object `tfsdk:"ddos"`
	WAF        types.Object `tfsdk:"waf"`
}

var pullzoneShieldDdosType = map[string]attr.Type{
	"level":            types.StringType,
	"mode":             types.StringType,
	"challenge_window": types.Int64Type,
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
				},
				Description: "Configures WAF settings.",
			},
		},
	}
}

func (r *PullzoneShieldResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
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
		}

		obj, diags := types.ObjectValue(pullzoneShieldWafType, values)
		if diags != nil {
			return PullzoneShieldResourceModel{}, diags
		}

		dataTf.WAF = obj
	}

	return dataTf, nil
}
