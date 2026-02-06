// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/resourcestateupgrader"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"regexp"
	"strconv"
	"strings"
)

var _ resource.Resource = &PullzoneRatelimitRuleResource{}
var _ resource.ResourceWithImportState = &PullzoneRatelimitRuleResource{}
var _ resource.ResourceWithUpgradeState = &PullzoneRatelimitRuleResource{}

func NewPullzoneRatelimitRule() resource.Resource {
	return &PullzoneRatelimitRuleResource{}
}

type PullzoneRatelimitRuleResource struct {
	client *api.Client
}

type PullzoneRatelimitRuleResourceModel struct {
	Id              types.Int64  `tfsdk:"id"`
	PullzoneId      types.Int64  `tfsdk:"pullzone"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Conditions      types.List   `tfsdk:"condition"`
	Transformations types.Set    `tfsdk:"transformations"`
	Limit           types.Object `tfsdk:"limit"`
	Response        types.Object `tfsdk:"response"`
}

var pullzoneRatelimitConditionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"variable":       types.StringType,
		"variable_value": types.StringType,
		"operator":       types.StringType,
		"value":          types.StringType,
	},
}

var pullzoneRatelimitRuleLimitType = map[string]attr.Type{
	"requests": types.Int64Type,
	"interval": types.Int64Type,
}

var pullzoneRatelimitRuleResponseType = map[string]attr.Type{
	"interval": types.Int64Type,
}

func (r *PullzoneRatelimitRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_ratelimit_rule"
}

func (r *PullzoneRatelimitRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
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
			"transformations": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(maps.Values(pullzoneShieldRuleTransformationMap)...),
					),
				},
				Description: generateMarkdownMapOptions(pullzoneShieldRuleTransformationMap),
			},
		},
		Blocks: map[string]schema.Block{
			"condition": schema.ListNestedBlock{
				Description: "The condition to trigger the rate limit rule.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"variable": schema.StringAttribute{
							// @TODO some variables are only available on advanced plan
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(maps.Values(pullzoneShieldRuleConditionVariableMap)...),
							},
							Description: generateMarkdownMapOptions(pullzoneShieldRuleConditionVariableMap),
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
								stringvalidator.OneOf(maps.Values(pullzoneShieldRuleConditionOperationMap)...),
							},
							Description: generateMarkdownMapOptions(pullzoneShieldRuleConditionOperationMap),
						},
						"value": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Validators: []validator.Object{
						objectvalidator.IsRequired(),
					},
					PlanModifiers: []planmodifier.Object{
						objectplanmodifier.UseStateForUnknown(),
					},
				},
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
							int64validator.OneOf(pullzoneShieldRatelimitRuleLimitTimeframeOptions...),
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
							int64validator.OneOf(pullzoneShieldRatelimitRuleResponseTimeframeOptions...),
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

func (r *PullzoneRatelimitRuleResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: resourcestateupgrader.PullzoneRatelimitRuleV0},
	}
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

	dataApi, err := r.client.GetPullzoneRatelimitRule(ctx, data.PullzoneId.ValueInt64(), data.Id.ValueInt64())
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
	dataApiResult, err := r.client.UpdatePullzoneRatelimitRule(ctx, dataApi)
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

	err := r.client.DeletePullzoneRatelimitRule(ctx, data.Id.ValueInt64())
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

	dataApi, err := r.client.GetPullzoneRatelimitRule(ctx, pullzoneId, ruleId)
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

	// conditions
	{
		conditions := dataTf.Conditions.Elements()

		for i, c := range conditions {
			conditionAttr := c.(types.Object).Attributes()

			variable := conditionAttr["variable"].(types.String).ValueString()
			variableValue := conditionAttr["variable_value"].(types.String).ValueString()
			variableTypes := map[string]string{variable: variableValue}
			operator := mapValueToKey(pullzoneShieldRuleConditionOperationMap, conditionAttr["operator"].(types.String).ValueString())
			value := conditionAttr["value"].(types.String).ValueString()

			if i == 0 {
				dataApi.RuleConfiguration.VariableTypes = variableTypes
				dataApi.RuleConfiguration.OperatorType = operator
				dataApi.RuleConfiguration.Value = value
			} else {
				dataApi.RuleConfiguration.ChainedRules = append(dataApi.RuleConfiguration.ChainedRules, api.PullzoneRatelimitRuleChainedRule{
					VariableTypes: variableTypes,
					OperatorType:  operator,
					Value:         value,
				})
			}
		}
	}

	// transformations
	{
		transformationElements := dataTf.Transformations.Elements()
		transformationIds := make([]int64, 0, len(transformationElements))

		for _, t := range transformationElements {
			v := mapValueToKey(pullzoneShieldRuleTransformationMap, t.(types.String).ValueString())
			transformationIds = append(transformationIds, v)
		}

		dataApi.RuleConfiguration.TransformationTypes = transformationIds
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
	var diags diag.Diagnostics

	dataTf := PullzoneRatelimitRuleResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Description = types.StringValue(dataApi.Description)

	conditions := make([]attr.Value, 0, len(dataApi.RuleConfiguration.ChainedRules)+1)

	// condition
	{
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

		variableMapByValue := utils.MapInvert(pullzoneShieldRuleConditionVariableMap)
		if _, ok := variableMapByValue[variable]; !ok {
			diags.AddError("Invalid API response", fmt.Sprintf("The API returned a variable that is not supported by the provider: %s", variable))
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		conditionObj, diags := types.ObjectValue(pullzoneRatelimitConditionType.AttrTypes, map[string]attr.Value{
			"operator":       types.StringValue(mapKeyToValue(pullzoneShieldRuleConditionOperationMap, dataApi.RuleConfiguration.OperatorType)),
			"value":          types.StringValue(dataApi.RuleConfiguration.Value),
			"variable":       types.StringValue(variable),
			"variable_value": variableValue,
		})

		if diags.HasError() {
			return PullzoneRatelimitRuleResourceModel{}, diags
		}

		conditions = append(conditions, conditionObj)
	}

	// chained conditions
	{
		for _, rule := range dataApi.RuleConfiguration.ChainedRules {
			var variable string
			var variableValue types.String

			for k, v := range rule.VariableTypes {
				variable = k
				if v == "" {
					variableValue = types.StringNull()
				} else {
					variableValue = types.StringValue(v)
				}

				break
			}

			condition, diags := types.ObjectValue(pullzoneRatelimitConditionType.AttrTypes, map[string]attr.Value{
				"operator":       types.StringValue(mapKeyToValue(pullzoneShieldRuleConditionOperationMap, rule.OperatorType)),
				"variable":       types.StringValue(variable),
				"variable_value": variableValue,
				"value":          types.StringValue(rule.Value),
			})

			if diags.HasError() {
				return dataTf, diags
			}

			conditions = append(conditions, condition)
		}
	}

	slices.SortFunc(conditions, func(a, b attr.Value) int {
		var aValue string
		var bValue string

		aAttr := a.(types.Object).Attributes()
		bAttr := b.(types.Object).Attributes()

		if v, ok := aAttr["variable"]; ok {
			aValue = v.(types.String).ValueString()
		}

		if v, ok := bAttr["variable"]; ok {
			bValue = v.(types.String).ValueString()
		}

		return strings.Compare(aValue, bValue)
	})

	conditionList, diags := types.ListValue(pullzoneRatelimitConditionType, conditions)
	if diags.HasError() {
		return dataTf, diags
	}

	dataTf.Conditions = conditionList

	// transformations
	{
		if len(dataApi.RuleConfiguration.TransformationTypes) == 0 {
			dataTf.Transformations = types.SetNull(types.StringType)
		} else {
			transformationValues := make([]attr.Value, 0, len(dataApi.RuleConfiguration.TransformationTypes))

			for _, r := range dataApi.RuleConfiguration.TransformationTypes {
				value := mapKeyToValue(pullzoneShieldRuleTransformationMap, r)
				transformationValues = append(transformationValues, types.StringValue(value))
			}

			transformationsSet, diags := types.SetValue(types.StringType, transformationValues)
			if diags.HasError() {
				return dataTf, diags
			}

			dataTf.Transformations = transformationsSet
		}
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
