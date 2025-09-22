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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"strings"
)

var _ resource.Resource = &DnsScriptVariableResource{}
var _ resource.ResourceWithImportState = &DnsScriptVariableResource{}

func NewDnsScriptVariableResource() resource.Resource {
	return &DnsScriptVariableResource{}
}

type DnsScriptVariableResource struct {
	client *api.Client
}

type DnsScriptVariableResourceModel struct {
	Id     types.Int64  `tfsdk:"id"`
	Script types.Int64  `tfsdk:"script"`
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
}

func (r *DnsScriptVariableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_script_variable"
}

func (r *DnsScriptVariableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages an Environment Variable for a DNS Script in bunny.net.",

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
				Description: "The ID of the associated DNS script.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the environment variable.",
			},
			"value": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The value of the environment variable.",
			},
		},
	}
}

func (r *DnsScriptVariableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DnsScriptVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf DnsScriptVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)

	script, err := r.client.GetComputeScript(ctx, dataApi.ScriptId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DNS script variable", err.Error())
		return
	}

	if script.ScriptType != api.ScriptTypeDns {
		resp.Diagnostics.AddError("Unable to create DNS script variable", "404 Not Found")
		return
	}

	dataApi, err = r.client.CreateComputeScriptVariable(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DNS script variable", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created DNS script variable %d", dataApi.Id))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DnsScriptVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetComputeScriptVariable(data.Script.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching DNS script variable", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DnsScriptVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateComputeScriptVariable(dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating DNS script variable", err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated DNS script variable %d", dataApi.Id))

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DnsScriptVariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeScriptVariable(data.Script.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting DNS script variable", err.Error()))
	}
}

func (r *DnsScriptVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scriptIdStr, name, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding DNS script variable", "Use \"<scriptId>|<variableName>\" as ID on terraform import command"))
		return
	}

	scriptId, err := strconv.ParseInt(scriptIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding DNS script variable", "Invalid DNS script ID: "+err.Error()))
		return
	}

	script, err := r.client.GetComputeScript(ctx, scriptId)
	if err != nil {
		resp.Diagnostics.AddError("Error finding DNS script variable", err.Error())
		return
	}

	if script.ScriptType != api.ScriptTypeDns {
		resp.Diagnostics.AddError("Error finding DNS script variable", "404 Not Found")
		return
	}

	dataApi, err := r.client.GetComputeScriptVariableByName(scriptId, name)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching DNS script variable", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptVariableResource) convertModelToApi(ctx context.Context, dataTf DnsScriptVariableResourceModel) api.ComputeScriptVariable {
	dataApi := api.ComputeScriptVariable{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.ScriptId = dataTf.Script.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.DefaultValue = dataTf.Value.ValueString()

	return dataApi
}

func (r *DnsScriptVariableResource) convertApiToModel(dataApi api.ComputeScriptVariable) (DnsScriptVariableResourceModel, diag.Diagnostics) {
	dataTf := DnsScriptVariableResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Script = types.Int64Value(dataApi.ScriptId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Value = types.StringValue(dataApi.DefaultValue)

	return dataTf, nil
}
