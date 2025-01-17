// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"strconv"
)

var _ resource.Resource = &ComputeScriptResource{}
var _ resource.ResourceWithImportState = &ComputeScriptResource{}

func NewComputeScriptResource() resource.Resource {
	return &ComputeScriptResource{}
}

type ComputeScriptResource struct {
	client *api.Client
}

type ComputeScriptResourceModel struct {
	Id            types.Int64  `tfsdk:"id"`
	Type          types.String `tfsdk:"type"`
	Name          types.String `tfsdk:"name"`
	Content       types.String `tfsdk:"content"`
	DeploymentKey types.String `tfsdk:"deployment_key"`
	Release       types.String `tfsdk:"release"`
}

var computeScriptTypeMap = map[uint8]string{
	1: "standalone",
	2: "middleware",
}

func (r *ComputeScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_script"
}

func (r *ComputeScriptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a compute script in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the script.",
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(computeScriptTypeMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(computeScriptTypeMap),
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the script.",
			},
			"content": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The code of the script.",
			},
			"deployment_key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The deployment key for the script.",
			},
			"release": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The current release identifier for the script.",
			},
		},
	}
}

func (r *ComputeScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ComputeScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf ComputeScriptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateComputeScript(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create compute script", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created compute script %d", dataApi.Id))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeScriptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetComputeScript(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching compute script", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ComputeScriptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)

	previousDataApi, err := r.client.GetComputeScript(dataApi.Id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching compute script", err.Error()))
		return
	}

	dataApiResult, err := r.client.UpdateComputeScript(dataApi, previousDataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating compute script", err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated compute script %d", dataApiResult.Id))

	dataTf, diags := r.convertApiToModel(dataApiResult)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComputeScriptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeScript(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting compute script", err.Error()))
	}
}

func (r *ComputeScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding compute script", "Invalid ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetComputeScript(id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching compute script", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeScriptResource) convertModelToApi(ctx context.Context, dataTf ComputeScriptResourceModel) api.ComputeScript {
	dataApi := api.ComputeScript{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.ScriptType = mapValueToKey(computeScriptTypeMap, dataTf.Type.ValueString())
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Content = dataTf.Content.ValueString()

	return dataApi
}

func (r *ComputeScriptResource) convertApiToModel(dataApi api.ComputeScript) (ComputeScriptResourceModel, diag.Diagnostics) {
	dataTf := ComputeScriptResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Type = types.StringValue(mapKeyToValue(computeScriptTypeMap, dataApi.ScriptType))
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Content = types.StringValue(dataApi.Content)
	dataTf.DeploymentKey = types.StringValue(dataApi.DeploymentKey)

	if dataApi.Release != "" {
		dataTf.Release = types.StringValue(dataApi.Release)
	} else {
		dataTf.Release = types.StringNull()
	}

	return dataTf, nil
}
