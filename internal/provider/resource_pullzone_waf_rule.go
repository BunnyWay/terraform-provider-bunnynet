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

var _ resource.Resource = &PullzoneWafRuleResource{}
var _ resource.ResourceWithImportState = &PullzoneWafRuleResource{}

func NewPullzoneWafRule() resource.Resource {
	return &PullzoneWafRuleResource{}
}

type PullzoneWafRuleResource struct {
	client *api.Client
}

type PullzoneWafRuleResourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	PullzoneId  types.Int64  `tfsdk:"pullzone"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Condition   types.Object `tfsdk:"condition"`
	Response    types.Object `tfsdk:"response"`
}

var pullzoneWafConditionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"variable":        types.StringType,
		"variable_value":  types.StringType,
		"operator":        types.StringType,
		"value":           types.StringType,
		"transformations": types.SetType{ElemType: types.StringType},
	},
}

var pullzoneWafRuleResponseType = map[string]attr.Type{
	"action": types.StringType,
}

// jq -r '.enumValues[] | (.value|tostring)+":"+.name'
// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WafRuleOperatorType")'
var pullzoneWafRuleConditionOperationMap = map[int64]string{
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
var pullzoneWafRuleConditionTransformationMap = map[int64]string{
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
var pullzoneWafRuleConditionVariableMap = map[int64]string{
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

// curl -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/waf/enums | jq -r '.data[] | select(.enumName=="WafRuleActionType")'
var pullzoneWafRuleResponseActionMap = map[uint8]string{
	1: "Block",
	2: "Log",
	3: "Challenge",
	4: "Allow",
	5: "Bypass",
}

func (r *PullzoneWafRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_waf_rule"
}

func (r *PullzoneWafRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a WAF rule for a bunny.net pullzone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the WAF rule.",
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
				Description: "The WAF rule name.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^([a-zA-Z0-9\ ]+)$`), "Only letters and numbers."),
				},
				Description: "The WAF rule description.",
			},
		},
		Blocks: map[string]schema.Block{
			"condition": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"variable": schema.StringAttribute{
						// @TODO some variables are only available on advanced
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneWafRuleConditionVariableMap)...),
						},
						Description: generateMarkdownMapOptions(pullzoneWafRuleConditionVariableMap),
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
							stringvalidator.OneOf(maps.Values(pullzoneWafRuleConditionOperationMap)...),
						},
						Description: generateMarkdownMapOptions(pullzoneWafRuleConditionOperationMap),
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
								stringvalidator.OneOf(maps.Values(pullzoneWafRuleConditionTransformationMap)...),
							),
						},
						Description: generateMarkdownMapOptions(pullzoneWafRuleConditionTransformationMap),
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Description: "The condition to trigger the WAF rule.",
			},
			"response": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"action": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(maps.Values(pullzoneWafRuleResponseActionMap)...),
						},
						Description: "The action to take if the WAF rule is triggered. " + generateMarkdownMapOptions(pullzoneWafRuleResponseActionMap),
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Description: "The response once the WAF rule is triggered.",
			},
		},
	}
}

func (r *PullzoneWafRuleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *PullzoneWafRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneWafRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneWafRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := dataTf.PullzoneId.ValueInt64()
	pzWafRuleMutex.Lock(pullzoneId)

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneWafRule(ctx, dataApi)

	pzWafRuleMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.AddError("Unable to create waf rule", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created waf rule for pullzone %d", pullzoneId))
	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneWafRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneWafRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetPullzoneWafRule(data.PullzoneId.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching waf rule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneWafRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneWafRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApiResult, err := r.client.UpdatePullzoneWafRule(dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating waf rule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApiResult)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneWafRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneWafRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	pzWafRuleMutex.Lock(pullzoneId)

	err := r.client.DeletePullzoneWafRule(data.Id.ValueInt64())

	pzWafRuleMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting waf rule", err.Error()))
	}
}

func (r *PullzoneWafRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	dataApi, err := r.client.GetPullzoneWafRule(pullzoneId, ruleId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching waf rule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneWafRuleResource) convertModelToApi(ctx context.Context, dataTf PullzoneWafRuleResourceModel) api.PullzoneWafRule {
	dataApi := api.PullzoneWafRule{}
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
		dataApi.RuleConfiguration.OperatorType = mapValueToKey(pullzoneWafRuleConditionOperationMap, conditionAttr["operator"].(types.String).ValueString())
		dataApi.RuleConfiguration.Value = conditionAttr["value"].(types.String).ValueString()

		for _, item := range conditionAttr["transformations"].(types.Set).Elements() {
			v := mapValueToKey(pullzoneWafRuleConditionTransformationMap, item.(types.String).ValueString())
			dataApi.RuleConfiguration.TransformationTypes = append(dataApi.RuleConfiguration.TransformationTypes, v)
		}
	}

	// response
	{
		attrs := dataTf.Response.Attributes()
		dataApi.RuleConfiguration.ActionType = mapValueToKey(pullzoneWafRuleResponseActionMap, attrs["action"].(types.String).ValueString())
	}

	return dataApi
}

func (r *PullzoneWafRuleResource) convertApiToModel(ctx context.Context, dataApi api.PullzoneWafRule) (PullzoneWafRuleResourceModel, diag.Diagnostics) {
	dataTf := PullzoneWafRuleResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Description = types.StringValue(dataApi.Description)

	// condition
	{
		var transformations []attr.Value
		for _, item := range dataApi.RuleConfiguration.TransformationTypes {
			transformations = append(transformations, types.StringValue(mapKeyToValue(pullzoneWafRuleConditionTransformationMap, item)))
		}

		transformationsSet, diags := types.SetValue(types.StringType, transformations)

		if diags.HasError() {
			return PullzoneWafRuleResourceModel{}, diags
		}

		if len(dataApi.RuleConfiguration.VariableTypes) != 1 {
			diags.AddError("Invalid API response", "The API returned multiple variables, but that's not supported by the provider.")
			return PullzoneWafRuleResourceModel{}, diags
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

		variableMapByValue := utils.MapInvert(pullzoneWafRuleConditionVariableMap)
		if _, ok := variableMapByValue[variable]; !ok {
			diags.AddError("Invalid API response", fmt.Sprintf("The API returned a variable that is not supported by the provider: %s", variable))
			return PullzoneWafRuleResourceModel{}, diags
		}

		condition, diags := types.ObjectValue(pullzoneWafConditionType.AttrTypes, map[string]attr.Value{
			"variable":        types.StringValue(variable),
			"variable_value":  variableValue,
			"operator":        types.StringValue(mapKeyToValue(pullzoneWafRuleConditionOperationMap, dataApi.RuleConfiguration.OperatorType)),
			"value":           types.StringValue(dataApi.RuleConfiguration.Value),
			"transformations": transformationsSet,
		})

		if diags.HasError() {
			return PullzoneWafRuleResourceModel{}, diags
		}

		dataTf.Condition = condition
	}

	// response
	{
		response, diags := types.ObjectValue(pullzoneWafRuleResponseType, map[string]attr.Value{
			"action": types.StringValue(mapKeyToValue(pullzoneWafRuleResponseActionMap, dataApi.RuleConfiguration.ActionType)),
		})

		if diags.HasError() {
			return PullzoneWafRuleResourceModel{}, diags
		}

		dataTf.Response = response
	}

	return dataTf, nil
}
