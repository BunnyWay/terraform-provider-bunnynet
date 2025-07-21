// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/pullzoneedgeruleresourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"strconv"
	"strings"
)

var _ resource.Resource = &PullzoneEdgeruleResource{}
var _ resource.ResourceWithImportState = &PullzoneEdgeruleResource{}

func NewPullzoneEdgeruleResource() resource.Resource {
	return &PullzoneEdgeruleResource{}
}

type PullzoneEdgeruleResource struct {
	client *api.Client
}

type PullzoneEdgeruleResourceModel struct {
	Id               types.String `tfsdk:"id"`
	PullzoneId       types.Int64  `tfsdk:"pullzone"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Description      types.String `tfsdk:"description"`
	Action           types.String `tfsdk:"action"`
	ActionParameter1 types.String `tfsdk:"action_parameter1"`
	ActionParameter2 types.String `tfsdk:"action_parameter2"`
	ActionParameter3 types.String `tfsdk:"action_parameter3"`
	MatchType        types.String `tfsdk:"match_type"`
	Priority         types.Int64  `tfsdk:"priority"`
	Actions          types.List   `tfsdk:"actions"`
	Triggers         types.List   `tfsdk:"triggers"`
}

func (r *PullzoneEdgeruleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_edgerule"
}

var pullzoneEdgeruleActionType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type":       types.StringType,
		"parameter1": types.StringType,
		"parameter2": types.StringType,
		"parameter3": types.StringType,
	},
}

var pullzoneEdgeruleTriggerType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type":       types.StringType,
		"match_type": types.StringType,
		"parameter1": types.StringType,
		"parameter2": types.StringType,
		"patterns": types.ListType{
			ElemType: types.StringType,
		},
	},
}

func (r *PullzoneEdgeruleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages edge rules for a bunny.net pull zone. It is used to define and configure rules that determine how content is delivered at the edge.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique GUID of the edge rule.",
			},
			"pullzone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"enabled": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the edge rule is enabled.",
			},
			"action": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneedgeruleresourcevalidator.ActionMap)...),
					stringvalidator.ConflictsWith(path.MatchRoot("actions")),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneedgeruleresourcevalidator.ActionMap),
			},
			"action_parameter1": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"action_parameter2": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"action_parameter3": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The description of the edge rule.",
			},
			"match_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("MatchAny"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneedgeruleresourcevalidator.TriggerMatchTypeMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneedgeruleresourcevalidator.TriggerMatchTypeMap),
			},
			"priority": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The priority of the edge rule. The lower number is executed first.",
			},
			"actions": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: pullzoneEdgeruleActionType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ConflictsWith(
						path.MatchRoot("action"),
						path.MatchRoot("action_parameter1"),
						path.MatchRoot("action_parameter2"),
					),
				},
				Description: "List of actions for the edge rule.",
			},
			"triggers": schema.ListAttribute{
				Required:    true,
				ElementType: pullzoneEdgeruleTriggerType,
			},
		},
	}
}

func (r *PullzoneEdgeruleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		pullzoneedgeruleresourcevalidator.TriggerObject(),
		pullzoneedgeruleresourcevalidator.ActionParameters(),
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("action"),
			path.MatchRoot("actions"),
		),
	}
}

func (r *PullzoneEdgeruleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneEdgeruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneEdgeruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := dataTf.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneEdgerule(dataApi)
	pzMutex.Unlock(pullzoneId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create edgerule", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created edgerule for pullzone %d", dataApi.PullzoneId))
	dataTf, diags := r.convertApiToModel(dataApi, !dataTf.Action.IsNull())
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneEdgeruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneEdgeruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	dataApi, err := r.client.GetPullzoneEdgerule(data.PullzoneId.ValueInt64(), data.Id.ValueString())
	pzMutex.Unlock(pullzoneId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching edgerule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi, !data.Action.IsNull())
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneEdgeruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneEdgeruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.CreatePullzoneEdgerule(dataApi)
	pzMutex.Unlock(pullzoneId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating edgerule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi, !data.Action.IsNull())
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneEdgeruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneEdgeruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	err := r.client.DeletePullzoneEdgerule(data.PullzoneId.ValueInt64(), data.Id.ValueString())
	pzMutex.Unlock(pullzoneId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting edgerule", err.Error()))
	}
}

func (r *PullzoneEdgeruleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pullzoneIdStr, guid, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding edgerule", "Use \"<pullzoneId>|<edgeruleGuid>\" as ID on terraform import command"))
		return
	}

	pullzoneId, err := strconv.ParseInt(pullzoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding edgerule", "Invalid pullzone ID: "+err.Error()))
		return
	}

	edgerule, err := r.client.GetPullzoneEdgerule(pullzoneId, guid)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding edgerule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(edgerule, false)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneEdgeruleResource) convertModelToApi(ctx context.Context, dataTf PullzoneEdgeruleResourceModel) api.PullzoneEdgerule {
	dataApi := api.PullzoneEdgerule{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.PullzoneId = dataTf.PullzoneId.ValueInt64()
	dataApi.Enabled = dataTf.Enabled.ValueBool()
	dataApi.Description = dataTf.Description.ValueString()
	dataApi.MatchType = mapValueToKey(pullzoneedgeruleresourcevalidator.TriggerMatchTypeMap, dataTf.MatchType.ValueString())
	dataApi.OrderIndex = dataTf.Priority.ValueInt64()

	// actions
	{
		actionsElements := dataTf.Actions.Elements()

		// action field
		if len(actionsElements) == 0 {
			dataApi.Action = mapValueToKey(pullzoneedgeruleresourcevalidator.ActionMap, dataTf.Action.ValueString())
			dataApi.ActionParameter1 = dataTf.ActionParameter1.ValueString()
			dataApi.ActionParameter2 = dataTf.ActionParameter2.ValueString()
			dataApi.ActionParameter3 = dataTf.ActionParameter3.ValueString()
			dataApi.ExtraActions = []api.PullzoneEdgeruleExtraAction{}
		} else {
			// actions list
			actions := make([]api.PullzoneEdgeruleExtraAction, len(actionsElements))
			for i, el := range actionsElements {
				action := el.(types.Object).Attributes()
				parameter1 := ""
				if t, ok := action["parameter1"]; ok && !t.(types.String).IsNull() {
					parameter1 = t.(types.String).ValueString()
				}

				parameter2 := ""
				if t, ok := action["parameter2"]; ok && !t.(types.String).IsNull() {
					parameter2 = t.(types.String).ValueString()
				}

				parameter3 := ""
				if t, ok := action["parameter3"]; ok && !t.(types.String).IsNull() {
					parameter3 = t.(types.String).ValueString()
				}

				actions[i] = api.PullzoneEdgeruleExtraAction{
					ActionType:       mapValueToKey(pullzoneedgeruleresourcevalidator.ActionMap, action["type"].(types.String).ValueString()),
					ActionParameter1: parameter1,
					ActionParameter2: parameter2,
					ActionParameter3: parameter3,
				}
			}

			dataApi.Action = actions[0].ActionType
			dataApi.ActionParameter1 = actions[0].ActionParameter1
			dataApi.ActionParameter2 = actions[0].ActionParameter2
			dataApi.ActionParameter3 = actions[0].ActionParameter3

			if len(actions) > 1 {
				dataApi.ExtraActions = actions[1:]
			}
		}
	}

	// triggers
	{
		triggerElements := dataTf.Triggers.Elements()
		triggers := make([]api.PullzoneEdgeruleTrigger, len(triggerElements))
		for i, el := range triggerElements {
			trigger := el.(types.Object).Attributes()

			patternElements := trigger["patterns"].(types.List).Elements()
			patterns := make([]string, len(patternElements))
			for j, pattern := range patternElements {
				patterns[j] = pattern.(types.String).ValueString()
			}

			parameter1 := ""
			if t, ok := trigger["parameter1"]; ok && !t.(types.String).IsNull() {
				parameter1 = t.(types.String).ValueString()
			}

			parameter2 := ""
			if t, ok := trigger["parameter2"]; ok && !t.(types.String).IsNull() {
				parameter2 = t.(types.String).ValueString()
			}

			triggers[i] = api.PullzoneEdgeruleTrigger{
				Type:       mapValueToKey(pullzoneedgeruleresourcevalidator.TriggerTypeMap, trigger["type"].(types.String).ValueString()),
				MatchType:  mapValueToKey(pullzoneedgeruleresourcevalidator.TriggerMatchTypeMap, trigger["match_type"].(types.String).ValueString()),
				Patterns:   patterns,
				Parameter1: parameter1,
				Parameter2: parameter2,
			}
		}

		dataApi.Triggers = triggers
	}

	return dataApi
}

func (r *PullzoneEdgeruleResource) convertApiToModel(dataApi api.PullzoneEdgerule, useSingleAction bool) (PullzoneEdgeruleResourceModel, diag.Diagnostics) {
	dataTf := PullzoneEdgeruleResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Enabled = types.BoolValue(dataApi.Enabled)
	dataTf.Description = types.StringValue(dataApi.Description)
	dataTf.MatchType = types.StringValue(mapKeyToValue(pullzoneedgeruleresourcevalidator.TriggerMatchTypeMap, dataApi.MatchType))
	dataTf.Priority = types.Int64Value(dataApi.OrderIndex)

	// actions
	{
		i := 0
		actions := make([]attr.Value, len(dataApi.ExtraActions)+1)

		// main action
		{
			v, ok := pullzoneedgeruleresourcevalidator.ActionMap[dataApi.Action]
			if !ok {
				var diags diag.Diagnostics
				diags.AddError("Undefined Edge Rule action", fmt.Sprintf("Action %d is not defined in terraform", dataApi.Action))
				return PullzoneEdgeruleResourceModel{}, diags
			}

			actionValue, diags := types.ObjectValue(pullzoneEdgeruleActionType.AttrTypes, map[string]attr.Value{
				"type":       types.StringValue(v),
				"parameter1": typeStringOrNull(dataApi.ActionParameter1),
				"parameter2": typeStringOrNull(dataApi.ActionParameter2),
				"parameter3": typeStringOrNull(dataApi.ActionParameter3),
			})

			if diags != nil {
				return PullzoneEdgeruleResourceModel{}, diags
			}

			actions[i] = actionValue
			i++
		}

		// extra actions
		for _, extraAction := range dataApi.ExtraActions {
			v, ok := pullzoneedgeruleresourcevalidator.ActionMap[extraAction.ActionType]
			if !ok {
				var diags diag.Diagnostics
				diags.AddError("Undefined Edge Rule action", fmt.Sprintf("Action %d is not defined in terraform", dataApi.Action))
				return PullzoneEdgeruleResourceModel{}, diags
			}

			actionValue, diags := types.ObjectValue(pullzoneEdgeruleActionType.AttrTypes, map[string]attr.Value{
				"type":       types.StringValue(v),
				"parameter1": typeStringOrNull(extraAction.ActionParameter1),
				"parameter2": typeStringOrNull(extraAction.ActionParameter2),
				"parameter3": typeStringOrNull(extraAction.ActionParameter3),
			})

			if diags != nil {
				return PullzoneEdgeruleResourceModel{}, diags
			}

			actions[i] = actionValue
			i++
		}

		if len(actions) == 1 && useSingleAction {
			actionAttr := actions[0].(types.Object).Attributes()
			dataTf.Action = actionAttr["type"].(types.String)
			dataTf.ActionParameter1 = actionAttr["parameter1"].(types.String)
			dataTf.ActionParameter2 = actionAttr["parameter2"].(types.String)
			dataTf.ActionParameter3 = actionAttr["parameter3"].(types.String)
			dataTf.Actions = types.ListNull(pullzoneEdgeruleActionType)
		} else {
			actionsList, diags := types.ListValue(pullzoneEdgeruleActionType, actions)
			if diags != nil {
				return PullzoneEdgeruleResourceModel{}, diags
			}

			dataTf.Actions = actionsList
		}
	}

	// triggers
	{
		if len(dataApi.Triggers) == 0 {
			dataTf.Triggers = types.ListNull(types.ObjectType{})
		} else {
			triggers := make([]attr.Value, len(dataApi.Triggers))

			for i, tr := range dataApi.Triggers {
				patterns := make([]attr.Value, len(tr.Patterns))
				for j, value := range tr.Patterns {
					patterns[j] = types.StringValue(value)
				}

				patternsList, diags := types.ListValue(types.StringType, patterns)
				if diags != nil {
					return PullzoneEdgeruleResourceModel{}, diags
				}

				triggerValue, diags := types.ObjectValue(pullzoneEdgeruleTriggerType.AttrTypes, map[string]attr.Value{
					"type":       types.StringValue(mapKeyToValue(pullzoneedgeruleresourcevalidator.TriggerTypeMap, tr.Type)),
					"match_type": types.StringValue(mapKeyToValue(pullzoneedgeruleresourcevalidator.TriggerMatchTypeMap, tr.MatchType)),
					"patterns":   patternsList,
					"parameter1": typeStringOrNull(tr.Parameter1),
					"parameter2": typeStringOrNull(tr.Parameter2),
				})

				if diags != nil {
					return PullzoneEdgeruleResourceModel{}, diags
				}

				triggers[i] = triggerValue
			}

			triggersList, diags := types.ListValue(pullzoneEdgeruleTriggerType, triggers)
			if diags != nil {
				return PullzoneEdgeruleResourceModel{}, diags
			}

			dataTf.Triggers = triggersList
		}
	}

	return dataTf, nil
}
