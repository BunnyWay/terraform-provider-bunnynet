// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"regexp"
	"strconv"
	"strings"
)

var _ resource.Resource = &PullzoneRatelimitRuleResource{}
var _ resource.ResourceWithImportState = &PullzoneRatelimitRuleResource{}

func NewPullzoneRatelimitRule() resource.Resource {
	return &PullzoneRatelimitRuleResource{}
}

type PullzoneRatelimitRuleResource struct {
	client *api.Client
}

type PullzoneRatelimitRuleResourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	PullzoneId  types.Int64  `tfsdk:"pullzone"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Condition   types.Object `tfsdk:"condition"`
	Limit       types.Object `tfsdk:"limit"`
	Response    types.Object `tfsdk:"response"`
}

var pullzoneRatelimitConditionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"variable":        types.StringType,
		"variable_value":  types.StringType,
		"operator":        types.StringType,
		"value":           types.StringType,
		"transformations": types.SetType{ElemType: types.StringType},
	},
}

var pullzoneRatelimitRuleLimitType = map[string]attr.Type{
	"requests": types.Int64Type,
	"interval": types.Int64Type,
}

var pullzoneRatelimitRuleResponseType = map[string]attr.Type{
	"interval": types.Int64Type,
}

// jq -r '.enumValues[] | (.value|tostring)+":"+.name'
// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WafRuleOperatorType")'
var pullzoneRatelimitRuleConditionOperationMap = map[int64]string{
	0:  "BEGINSWITH",
	1:  "ENDSWITH",
	2:  "CONTAINS",
	3:  "CONTAINSWORD",
	4:  "STRMATCH",
	5:  "EQ",
	6:  "GE",
	7:  "GT",
	8:  "LE",
	9:  "LT",
	12: "WITHIN",
	14: "RX",
	15: "STREQ",
	17: "DETECTSQLI",
	18: "DETECTXSS",
}

// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WafRuleTransformationType")'
var pullzoneRatelimitRuleConditionTransformationMap = map[int64]string{
	1:  "CMDLINE",
	2:  "COMPRESSWHITESPACE",
	3:  "CSSDECODE",
	4:  "HEXENCODE",
	5:  "HTMLENTITYDECODE",
	6:  "JSDECODE",
	7:  "LENGTH",
	8:  "LOWERCASE",
	9:  "MD5",
	10: "NORMALIZEPATH",
	11: "NORMALISEPATH",
	12: "NORMALIZEPATHWIN",
	13: "NORMALISEPATHWIN",
	14: "REMOVECOMMENTS",
	15: "REMOVENULLS",
	16: "REMOVEWHITESPACE",
	17: "REPLACECOMMENTS",
	18: "SHA1",
	19: "URLDECODE",
	20: "URLDECODEUNI",
	21: "UTF8TOUNICODE",
}

// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WafRuleVariableType")'
var pullzoneRatelimitRuleConditionVariableMap = map[int64]string{
	0:  "REQUEST_URI",
	1:  "REQUEST_URI_RAW",
	2:  "ARGS",
	3:  "ARGS_COMBINED_SIZE",
	4:  "ARGS_GET",
	5:  "ARGS_GET_NAMES",
	6:  "ARGS_POST",
	7:  "ARGS_POST_NAMES",
	8:  "FILES_NAMES",
	10: "REMOTE_ADDR",
	11: "QUERY_STRING",
	12: "REQUEST_BASENAME",
	13: "REQUEST_BODY",
	14: "REQUEST_COOKIES_NAMES",
	15: "REQUEST_COOKIES",
	16: "REQUEST_FILENAME",
	17: "REQUEST_HEADERS_NAMES",
	18: "REQUEST_HEADERS",
	19: "REQUEST_LINE",
	20: "REQUEST_METHOD",
	21: "REQUEST_PROTOCOL",
	22: "RESPONSE_BODY",
	23: "RESPONSE_HEADERS",
	24: "RESPONSE_STATUS",
}

// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WafRateLimitTimeframeType")'
var pullzoneRatelimitRuleLimitTimeframeOptions = []int64{1, 10, 30, 60, 300, 900}

// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WafRatelimitBlockType")'
var pullzoneRatelimitRuleResponseTimeframeOptions = []int64{30, 60, 300, 900, 1800, 3600}

func (r *PullzoneRatelimitRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_ratelimit_rule"
}

func (r *PullzoneRatelimitRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a rate limit rule for a bunny.net pullzone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the rate limit rule.",
			},
			"pullzone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				Description: "The ID of the linked pullzone.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`^([a-zA-Z0-9\ ]+)$`), "Only letters and numbers."),
				},
				Description: "The rate limit rule name.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^([a-zA-Z0-9\ ]+)$`), "Only letters and numbers."),
				},
				Description: "The rate limit rule description.",
			},
		},
		Blocks: map[string]schema.Block{
			"condition": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"variable": schema.StringAttribute{
						// @TODO some variables are only available on advanced plan
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneRatelimitRuleConditionVariableMap)...),
						},
						Description: generateMarkdownMapOptions(pullzoneRatelimitRuleConditionVariableMap),
					},
					"variable_value": schema.StringAttribute{
						// @TODO validate, depends on variable
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"operator": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneRatelimitRuleConditionOperationMap)...),
						},
						Description: generateMarkdownMapOptions(pullzoneRatelimitRuleConditionOperationMap),
					},
					"value": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"transformations": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Set{
							setvalidator.ValueStringsAre(
								stringvalidator.OneOf(maps.Values(pullzoneRatelimitRuleConditionTransformationMap)...),
							),
						},
						Description: generateMarkdownMapOptions(pullzoneRatelimitRuleConditionTransformationMap),
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Description: "The condition to trigger the rate limit rule.",
			},
			"limit": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"requests": schema.Int64Attribute{
						Required: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
						Description: "The number of request within the interval to trigger the rate limit rule.",
					},
					"interval": schema.Int64Attribute{
						Required: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.OneOf(pullzoneRatelimitRuleLimitTimeframeOptions...),
						},
						Description: "The interval, in seconds, to consider for to trigger the rate limit rule.",
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"response": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"interval": schema.Int64Attribute{
						Required: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.OneOf(pullzoneRatelimitRuleResponseTimeframeOptions...),
						},
						Description: "The interval, in seconds, that the rate limit will apply.",
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Description: "The response once the rate limit rule is triggered.",
			},
		},
	}
}

func (r *PullzoneRatelimitRuleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *PullzoneRatelimitRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneRatelimitRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneRatelimitRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneRatelimitRule(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create ratelimit rule", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created ratelimit rule for pullzone %d", dataTf.PullzoneId.ValueInt64()))
	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneRatelimitRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneRatelimitRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetPullzoneRatelimitRule(data.PullzoneId.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching ratelimit rule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneRatelimitRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneRatelimitRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApiResult, err := r.client.UpdatePullzoneRatelimitRule(dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating ratelimit rule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApiResult)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneRatelimitRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneRatelimitRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePullzoneRatelimitRule(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting ratelimit rule", err.Error()))
	}
}

func (r *PullzoneRatelimitRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pullzoneIdStr, ruleIdStr, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Use \"<pullzoneId>|<ruleID>\" as ID on terraform import command"))
		return
	}

	pullzoneId, err := strconv.ParseInt(pullzoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Invalid pullzoneId: "+err.Error()))
		return
	}

	ruleId, err := strconv.ParseInt(ruleIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Invalid ruleId: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetPullzoneRatelimitRule(pullzoneId, ruleId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching rate limit rule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneRatelimitRuleResource) convertModelToApi(ctx context.Context, dataTf PullzoneRatelimitRuleResourceModel) api.PullzoneRatelimitRule {
	dataApi := api.PullzoneRatelimitRule{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.PullzoneId = dataTf.PullzoneId.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Description = dataTf.Description.ValueString()

	// condition
	{
		conditionAttr := dataTf.Condition.Attributes()

		variable := conditionAttr["variable"].(types.String).ValueString()
		variableValue := conditionAttr["variable_value"].(types.String).ValueString()
		dataApi.RuleConfiguration.VariableTypes = map[string]string{variable: variableValue}
		dataApi.RuleConfiguration.OperatorType = mapValueToKey(pullzoneRatelimitRuleConditionOperationMap, conditionAttr["operator"].(types.String).ValueString())
		dataApi.RuleConfiguration.Value = conditionAttr["value"].(types.String).ValueString()

		for _, item := range conditionAttr["transformations"].(types.Set).Elements() {
			v := mapValueToKey(pullzoneRatelimitRuleConditionTransformationMap, item.(types.String).ValueString())
			dataApi.RuleConfiguration.TransformationTypes = append(dataApi.RuleConfiguration.TransformationTypes, v)
		}
	}

	// limit
	{
		attrs := dataTf.Limit.Attributes()
		dataApi.RuleConfiguration.RequestCount = attrs["requests"].(types.Int64).ValueInt64()
		dataApi.RuleConfiguration.Timeframe = attrs["interval"].(types.Int64).ValueInt64()
	}

	// response
	{
		attrs := dataTf.Response.Attributes()
		dataApi.RuleConfiguration.BlockTime = attrs["interval"].(types.Int64).ValueInt64()
		dataApi.RuleConfiguration.ActionType = 1 // RateLimit
	}

	return dataApi
}

func (r *PullzoneRatelimitRuleResource) convertApiToModel(ctx context.Context, dataApi api.PullzoneRatelimitRule) (PullzoneRatelimitRuleResourceModel, diag.Diagnostics) {
	dataTf := PullzoneRatelimitRuleResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Description = types.StringValue(dataApi.Description)

	// condition
	{
		var transformations []attr.Value
		for _, item := range dataApi.RuleConfiguration.TransformationTypes {
			transformations = append(transformations, types.StringValue(mapKeyToValue(pullzoneRatelimitRuleConditionTransformationMap, item)))
		}

		transformationsSet, diags := types.SetValue(types.StringType, transformations)

		if diags.HasError() {
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		if len(dataApi.RuleConfiguration.VariableTypes) != 1 {
			diags.AddError("Invalid API response", "The API returned multiple variables, but that's not supported by the provider.")
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		var variable string
		var variableValue types.String

		for k, v := range dataApi.RuleConfiguration.VariableTypes {
			variable = k
			if v == "" {
				variableValue = types.StringNull()
			} else {
				variableValue = types.StringValue(v)
			}

			break
		}

		variableMapByValue := utils.MapInvert(pullzoneRatelimitRuleConditionVariableMap)
		if _, ok := variableMapByValue[variable]; !ok {
			diags.AddError("Invalid API response", fmt.Sprintf("The API returned a variable that is not supported by the provider: %s", variable))
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		condition, diags := types.ObjectValue(pullzoneRatelimitConditionType.AttrTypes, map[string]attr.Value{
			"variable":        types.StringValue(variable),
			"variable_value":  variableValue,
			"operator":        types.StringValue(mapKeyToValue(pullzoneRatelimitRuleConditionOperationMap, dataApi.RuleConfiguration.OperatorType)),
			"value":           types.StringValue(dataApi.RuleConfiguration.Value),
			"transformations": transformationsSet,
		})

		if diags.HasError() {
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		dataTf.Condition = condition
	}

	// limits
	{
		limit, diags := types.ObjectValue(pullzoneRatelimitRuleLimitType, map[string]attr.Value{
			"requests": types.Int64Value(dataApi.RuleConfiguration.RequestCount),
			"interval": types.Int64Value(dataApi.RuleConfiguration.Timeframe),
		})

		if diags.HasError() {
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		dataTf.Limit = limit
	}

	// response
	{
		response, diags := types.ObjectValue(pullzoneRatelimitRuleResponseType, map[string]attr.Value{
			"interval": types.Int64Value(dataApi.RuleConfiguration.BlockTime),
		})

		if diags.HasError() {
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		dataTf.Response = response
	}

	return dataTf, nil
}
