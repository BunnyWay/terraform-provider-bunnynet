// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	MatchType        types.String `tfsdk:"match_type"`
	Triggers         types.List   `tfsdk:"triggers"`
}

var pullzoneEdgeruleMatchTypeMap = map[uint8]string{
	0: "MatchAny",
	1: "MatchAll",
	2: "MatchNone",
}

var pullzoneEdgeruleActionMap = map[uint8]string{
	0:  "ForceSSL",
	1:  "Redirect",
	2:  "OriginUrl",
	3:  "OverrideCacheTime",
	4:  "BlockRequest",
	5:  "SetResponseHeader",
	6:  "SetRequestHeader",
	7:  "ForceDownload",
	8:  "DisableTokenAuthentication",
	9:  "EnableTokenAuthentication",
	10: "OverrideCacheTimePublic",
	11: "IgnoreQueryString",
	12: "DisableOptimizer",
	13: "ForceCompression",
	14: "SetStatusCode",
	15: "BypassPermaCache",
	16: "OverrideBrowserCacheTime",
	17: "OriginStorage",
	18: "SetNetworkRateLimit",
	19: "SetConnectionLimit",
	20: "SetRequestsPerSecondLimit",
}

var pullzoneEdgeruleTriggerTypeMap = map[uint8]string{
	0:  "Url",
	1:  "RequestHeader",
	2:  "ResponseHeader",
	3:  "UrlExtension",
	4:  "CountryCode",
	5:  "RemoteIP",
	6:  "UrlQueryString",
	7:  "RandomChance",
	8:  "StatusCode",
	9:  "RequestMethod",
	10: "CookieValue",
	11: "CountryStateCode",
}

func (r *PullzoneEdgeruleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_edgerule"
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
		Description: "This resource manages edge rules for a bunny.net pull zone. It is used to define and configure rules that determine how content is delivered at the edge, such as URL redirects, custom caching policies, and header manipulations.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique GUID of the edge rule",
			},
			"pullzone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Determines if the edge rule is currently enabled or not",
			},
			"action": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneEdgeruleActionMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneEdgeruleActionMap),
			},
			"action_parameter1": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"action_parameter2": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString(""),
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
				Description: "The description of the edge rule",
			},
			"match_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("MatchAny"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneEdgeruleMatchTypeMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(pullzoneEdgeruleMatchTypeMap),
			},
			"triggers": schema.ListAttribute{
				Required:    true,
				ElementType: pullzoneEdgeruleTriggerType,
			},
		},
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

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneEdgerule(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create edgerule", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created edgerule for pullzone %d", dataApi.PullzoneId))
	dataTf, diags := r.convertApiToModel(dataApi)
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

	dataApi, err := r.client.GetPullzoneEdgerule(data.PullzoneId.ValueInt64(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching edgerule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
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

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.CreatePullzoneEdgerule(dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating edgerule", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
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

	err := r.client.DeletePullzoneEdgerule(data.PullzoneId.ValueInt64(), data.Id.ValueString())
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

	dataTf, diags := r.convertApiToModel(edgerule)
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
	dataApi.Action = mapValueToKey(pullzoneEdgeruleActionMap, dataTf.Action.ValueString())
	dataApi.ActionParameter1 = dataTf.ActionParameter1.ValueString()
	dataApi.ActionParameter2 = dataTf.ActionParameter2.ValueString()
	dataApi.Description = dataTf.Description.ValueString()
	dataApi.MatchType = mapValueToKey(pullzoneEdgeruleMatchTypeMap, dataTf.MatchType.ValueString())

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
			Type:       mapValueToKey(pullzoneEdgeruleTriggerTypeMap, trigger["type"].(types.String).ValueString()),
			MatchType:  mapValueToKey(pullzoneEdgeruleMatchTypeMap, trigger["match_type"].(types.String).ValueString()),
			Patterns:   patterns,
			Parameter1: parameter1,
			Parameter2: parameter2,
		}
	}

	dataApi.Triggers = triggers

	return dataApi
}

func (r *PullzoneEdgeruleResource) convertApiToModel(dataApi api.PullzoneEdgerule) (PullzoneEdgeruleResourceModel, diag.Diagnostics) {
	dataTf := PullzoneEdgeruleResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Enabled = types.BoolValue(dataApi.Enabled)
	dataTf.Action = types.StringValue(mapKeyToValue(pullzoneEdgeruleActionMap, dataApi.Action))
	dataTf.ActionParameter1 = types.StringValue(dataApi.ActionParameter1)
	dataTf.ActionParameter2 = types.StringValue(dataApi.ActionParameter2)
	dataTf.Description = types.StringValue(dataApi.Description)
	dataTf.MatchType = types.StringValue(mapKeyToValue(pullzoneEdgeruleMatchTypeMap, dataApi.MatchType))

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

			var parameter1 attr.Value
			if tr.Parameter1 == "" {
				parameter1 = types.StringNull()
			} else {
				parameter1 = types.StringValue(tr.Parameter1)
			}

			var parameter2 attr.Value
			if tr.Parameter2 == "" {
				parameter2 = types.StringNull()
			} else {
				parameter2 = types.StringValue(tr.Parameter2)
			}

			triggerValue, diags := types.ObjectValue(pullzoneEdgeruleTriggerType.AttrTypes, map[string]attr.Value{
				"type":       types.StringValue(mapKeyToValue(pullzoneEdgeruleTriggerTypeMap, tr.Type)),
				"match_type": types.StringValue(mapKeyToValue(pullzoneEdgeruleMatchTypeMap, tr.MatchType)),
				"patterns":   patternsList,
				"parameter1": parameter1,
				"parameter2": parameter2,
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

	return dataTf, nil
}
