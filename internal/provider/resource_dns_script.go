// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
)

var _ resource.Resource = &DnsScriptResource{}
var _ resource.ResourceWithImportState = &DnsScriptResource{}
var _ resource.ResourceWithModifyPlan = &DnsScriptResource{}

func NewDnsScriptResource() resource.Resource {
	return &DnsScriptResource{}
}

type DnsScriptResource struct {
	client *api.Client
}

type DnsScriptResourceModel struct {
	Id      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
	Release types.String `tfsdk:"release"`
}

func (r *DnsScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_script"
}

func (r *DnsScriptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a DNS script in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the script.",
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

func (r *DnsScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DnsScriptResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var contentState string
	var contentPlan string

	req.State.GetAttribute(ctx, path.Root("content"), &contentState)
	req.Plan.GetAttribute(ctx, path.Root("content"), &contentPlan)

	if contentState == "" || contentPlan == "" || contentState == contentPlan {
		return
	}

	resp.Plan.SetAttribute(ctx, path.Root("release"), types.StringUnknown())
}

func (r *DnsScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf DnsScriptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateComputeScript(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DNS script", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created DNS script %d", dataApi.Id))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DnsScriptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetComputeScript(ctx, data.Id.ValueInt64())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching DNS script", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DnsScriptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)

	previousDataApi, err := r.client.GetComputeScript(ctx, dataApi.Id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching DNS script", err.Error()))
		return
	}

	dataApiResult, err := r.client.UpdateDnsScript(ctx, dataApi, previousDataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating DNS script", err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated DNS script %d", dataApiResult.Id))

	dataTf, diags := r.convertApiToModel(dataApiResult)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DnsScriptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeScript(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting DNS script", err.Error()))
	}
}

func (r *DnsScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding DNS script", "Invalid ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetComputeScript(ctx, id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching DNS script", err.Error()))
		return
	}

	if dataApi.ScriptType != api.ScriptTypeDns {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching DNS script", "404 Not Found"))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsScriptResource) convertModelToApi(ctx context.Context, dataTf DnsScriptResourceModel) api.ComputeScript {
	dataApi := api.ComputeScript{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Content = dataTf.Content.ValueString()
	dataApi.ScriptType = api.ScriptTypeDns

	return dataApi
}

func (r *DnsScriptResource) convertApiToModel(dataApi api.ComputeScript) (DnsScriptResourceModel, diag.Diagnostics) {
	dataTf := DnsScriptResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Content = types.StringValue(dataApi.Content)

	if dataApi.Release != "" {
		dataTf.Release = types.StringValue(dataApi.Release)
	} else {
		dataTf.Release = types.StringNull()
	}

	return dataTf, nil
}
