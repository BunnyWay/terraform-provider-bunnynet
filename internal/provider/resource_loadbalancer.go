// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"strconv"
)

var _ resource.Resource = &LoadbalancerResource{}
var _ resource.ResourceWithConfigure = &LoadbalancerResource{}
var _ resource.ResourceWithImportState = &LoadbalancerResource{}

func NewLoadbalancerResource() resource.Resource {
	return &LoadbalancerResource{}
}

type LoadbalancerResource struct {
	client *api.Client
}

type LoadbalancerResourceModel struct {
	Id             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Method         types.String `tfsdk:"method"`
	StickySessions types.String `tfsdk:"sticky_sessions"`
}

func (r *LoadbalancerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer"
}

func (r *LoadbalancerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a Load Balancer in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique ID of the Load Balancer.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Load Balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"method": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Weight"),
				MarkdownDescription: generateMarkdownMapOptions(loadbalancerMethodMap),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(loadbalancerMethodMap)...),
				},
			},
			"sticky_sessions": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Off"),
				MarkdownDescription: generateMarkdownMapOptions(loadbalancerStickySessionMap),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(loadbalancerStickySessionMap)...),
				},
			},
		},
	}
}

func (r *LoadbalancerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LoadbalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf LoadbalancerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateLoadbalancer(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.AddError("Unable to create loadbalancer", err.Error())
		return
	}

	tflog.Trace(ctx, "created loadbalancer "+dataApi.Name)
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LoadbalancerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetLoadbalancer(ctx, data.Id.ValueInt64())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching loadbalancer", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LoadbalancerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateLoadbalancer(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating loadbalancer", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LoadbalancerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLoadbalancer(ctx, data.Id.ValueInt64())
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting loadbalancer", err.Error()))
	}
}

func (r *LoadbalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error converting ID to integer", err.Error()))
		return
	}

	dataApi, err := r.client.GetLoadbalancer(ctx, id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching loadbalancer", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerResource) convertModelToApi(ctx context.Context, dataTf LoadbalancerResourceModel) api.Loadbalancer {
	dataApi := api.Loadbalancer{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Method = mapValueToKey(loadbalancerMethodMap, dataTf.Method.ValueString())
	dataApi.StickySession = mapValueToKey(loadbalancerStickySessionMap, dataTf.StickySessions.ValueString())

	return dataApi
}

func (r *LoadbalancerResource) convertApiToModel(dataApi api.Loadbalancer) (LoadbalancerResourceModel, diag.Diagnostics) {
	dataTf := LoadbalancerResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Method = types.StringValue(mapKeyToValue(loadbalancerMethodMap, dataApi.Method))
	dataTf.StickySessions = types.StringValue(mapKeyToValue(loadbalancerStickySessionMap, dataApi.StickySession))

	return dataTf, nil
}
