// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
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

var _ resource.Resource = &ComputeScriptSecretResource{}
var _ resource.ResourceWithImportState = &ComputeScriptSecretResource{}

func NewComputeScriptSecretResource() resource.Resource {
	return &ComputeScriptSecretResource{}
}

type ComputeScriptSecretResource struct {
	client *api.Client
}

type ComputeScriptSecretResourceModel struct {
	Id     types.Int64  `tfsdk:"id"`
	Script types.Int64  `tfsdk:"script"`
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
}

func (r *ComputeScriptSecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_script_secret"
}

func (r *ComputeScriptSecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a secret for a Compute Script in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the secret.",
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
				Description: "The name of the secret.",
			},
			"value": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The value of the secret.",
			},
		},
	}
}

func (r *ComputeScriptSecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ComputeScriptSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf ComputeScriptSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)

	script, err := r.client.GetComputeScript(ctx, dataApi.ScriptId)
	if err != nil {
		resp.Diagnostics.AddError("Error finding compute script secret", err.Error())
		return
	}

	if script.ScriptType == api.ScriptTypeDns {
		resp.Diagnostics.AddError("Error finding compute script secret", "404 Not Found")
		return
	}

	value := dataApi.Value
	dataApi, err = r.client.CreateComputeScriptSecret(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create compute script secret", err.Error())
		return
	}

	dataApi.Value = value
	tflog.Trace(ctx, fmt.Sprintf("created compute script secret %d", dataApi.Id))

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeScriptSecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetComputeScriptSecretByName(data.Script.ValueInt64(), data.Name.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching compute script secret", err.Error()))
		return
	}

	dataApi.Value = data.Value.ValueString()

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ComputeScriptSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateComputeScriptSecret(dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating compute script secret", err.Error()))
		return
	}

	dataApi.Value = data.Value.ValueString()
	tflog.Trace(ctx, fmt.Sprintf("updated compute script secret %d", dataApi.Id))

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComputeScriptSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeScriptSecret(data.Script.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting compute script secret", err.Error()))
	}
}

func (r *ComputeScriptSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scriptIdStr, name, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding compute script secret", "Use \"<scriptId>|<secretName>\" as ID on terraform import command"))
		return
	}

	scriptId, err := strconv.ParseInt(scriptIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding compute script secret", "Invalid compute script ID: "+err.Error()))
		return
	}

	script, err := r.client.GetComputeScript(ctx, scriptId)
	if err != nil {
		resp.Diagnostics.AddError("Error finding compute script secret", err.Error())
		return
	}

	if script.ScriptType == api.ScriptTypeDns {
		resp.Diagnostics.AddError("Error finding compute script secret", "404 Not Found")
		return
	}

	dataApi, err := r.client.GetComputeScriptSecretByName(scriptId, name)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching compute script secret", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptSecretResource) convertModelToApi(ctx context.Context, dataTf ComputeScriptSecretResourceModel) api.ComputeScriptSecret {
	dataApi := api.ComputeScriptSecret{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.ScriptId = dataTf.Script.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Value = dataTf.Value.ValueString()

	return dataApi
}

func (r *ComputeScriptSecretResource) convertApiToModel(dataApi api.ComputeScriptSecret) (ComputeScriptSecretResourceModel, diag.Diagnostics) {
	dataTf := ComputeScriptSecretResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Script = types.Int64Value(dataApi.ScriptId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Value = types.StringValue(dataApi.Value)

	return dataTf, nil
}
