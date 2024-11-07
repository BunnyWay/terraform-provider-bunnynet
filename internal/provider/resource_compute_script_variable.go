// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"strings"
)

var _ resource.Resource = &ComputeScriptVariableResource{}
var _ resource.ResourceWithImportState = &ComputeScriptVariableResource{}

func NewComputeScriptVariableResource() resource.Resource {
	return &ComputeScriptVariableResource{}
}

type ComputeScriptVariableResource struct {
	client *api.Client
}

type ComputeScriptVariableResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Script       types.Int64  `tfsdk:"script"`
	Name         types.String `tfsdk:"name"`
	DefaultValue types.String `tfsdk:"default_value"`
	Required     types.Bool   `tfsdk:"required"`
}

func (r *ComputeScriptVariableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_script_variable"
}

func (r *ComputeScriptVariableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages an Environment variable for a Compute Script in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the environment variable.",
			},
			"script": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: "The ID of the associated compute script.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the environment variable.",
			},
			"default_value": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The default value of the environment variable.",
			},
			"required": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the environment variable is required.",
			},
		},
	}
}

func (r *ComputeScriptVariableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ComputeScriptVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf ComputeScriptVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateComputeScriptVariable(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create compute script variable", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created compute script variable %d", dataApi.Id))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeScriptVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetComputeScriptVariable(data.Script.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching compute script variable", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ComputeScriptVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateComputeScriptVariable(dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating compute script variable", err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated compute script variable %d", dataApi.Id))

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComputeScriptVariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeScriptVariable(data.Script.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting compute script variable", err.Error()))
	}
}

func (r *ComputeScriptVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scriptIdStr, name, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding compute script variable", "Use \"<scriptId>|<variableName>\" as ID on terraform import command"))
		return
	}

	scriptId, err := strconv.ParseInt(scriptIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding compute script variable", "Invalid compute script ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetComputeScriptVariableByName(scriptId, name)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching compute script variable", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptVariableResource) convertModelToApi(ctx context.Context, dataTf ComputeScriptVariableResourceModel) api.ComputeScriptVariable {
	dataApi := api.ComputeScriptVariable{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.ScriptId = dataTf.Script.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.DefaultValue = dataTf.DefaultValue.ValueString()
	dataApi.Required = dataTf.Required.ValueBool()

	return dataApi
}

func (r *ComputeScriptVariableResource) convertApiToModel(dataApi api.ComputeScriptVariable) (ComputeScriptVariableResourceModel, diag.Diagnostics) {
	dataTf := ComputeScriptVariableResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Script = types.Int64Value(dataApi.ScriptId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.DefaultValue = types.StringValue(dataApi.DefaultValue)
	dataTf.Required = types.BoolValue(dataApi.Required)

	return dataTf, nil
}
