// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
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

var _ resource.Resource = &LoadbalancerGroupResource{}
var _ resource.ResourceWithConfigure = &LoadbalancerGroupResource{}
var _ resource.ResourceWithImportState = &LoadbalancerGroupResource{}

func NewLoadbalancerGroupResource() resource.Resource {
	return &LoadbalancerGroupResource{}
}

type LoadbalancerGroupResource struct {
	client *api.Client
}

type LoadbalancerGroupResourceModel struct {
	Id             types.Int64 `tfsdk:"id"`
	LoadbalancerId types.Int64 `tfsdk:"loadbalancer"`
	Priority       types.Int64 `tfsdk:"priority"`

	Enabled            types.Bool    `tfsdk:"enabled"`
	Name               types.String  `tfsdk:"name"`
	Method             types.String  `tfsdk:"method"`
	Latitude           types.Float64 `tfsdk:"latitude"`
	Longitude          types.Float64 `tfsdk:"longitude"`
	HealthcheckEnabled types.Bool    `tfsdk:"healthcheck_enabled"`
	HealthcheckPath    types.String  `tfsdk:"healthcheck_path"`
}

func (r *LoadbalancerGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_group"
}

func (r *LoadbalancerGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a Load Balancer group in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique ID of the Load Balancer group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"loadbalancer": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Load Balancer.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "The priority of the Load Balancer group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Indicates whether the Load Balancer group is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Load Balancer group.",
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
				MarkdownDescription: generateMarkdownMapOptions(loadbalancerGroupMethodMap),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(loadbalancerGroupMethodMap)...),
				},
			},
			"latitude": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(0),
				Description: `The latitude of the Load Balancer group. Used in Load Balancers with method = "Geo".`,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"longitude": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(0),
				Description: `The longitude of the Load Balancer group. Used for Load Balancers with method = "Geo"`,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"healthcheck_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Indicates whether healthcheck will be performed against all origins.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"healthcheck_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/"),
				Description: "The path that will be called to determine if an origin is healthy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
				},
			},
		},
	}
}

func (r *LoadbalancerGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LoadbalancerGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf LoadbalancerGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateLoadbalancerGroup(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.AddError("Unable to create loadbalancer group", err.Error())
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

func (r *LoadbalancerGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LoadbalancerGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetLoadbalancerGroup(ctx, data.LoadbalancerId.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching loadbalancer group", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LoadbalancerGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateLoadbalancerGroup(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating loadbalancer group", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LoadbalancerGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLoadbalancerGroup(ctx, data.LoadbalancerId.ValueInt64(), data.Id.ValueInt64())
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting loadbalancer group", err.Error()))
	}
}

func (r *LoadbalancerGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "|", 2)
	if len(parts) != 2 {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Use \"<loadbalancerId>|<groupName>\" as ID on terraform import command"))
		return
	}

	loadbalancerId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Invalid loadbalancerId: "+err.Error()))
		return
	}

	groupName := parts[1]

	loadbalancer, err := r.client.GetLoadbalancer(ctx, loadbalancerId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching loadbalancer", err.Error()))
		return
	}

	for _, group := range loadbalancer.Groups {
		if group.Name != groupName {
			continue
		}

		group.LoadbalancerId = loadbalancer.Id
		dataTf, diags := r.convertApiToModel(group)
		if diags != nil {
			resp.Diagnostics.Append(diags...)
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
		return
	}

	resp.Diagnostics.Append(diag.NewErrorDiagnostic("loadbalancer does have group", groupName))
}

func (r *LoadbalancerGroupResource) convertModelToApi(ctx context.Context, dataTf LoadbalancerGroupResourceModel) api.LoadbalancerGroup {
	dataApi := api.LoadbalancerGroup{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.LoadbalancerId = dataTf.LoadbalancerId.ValueInt64()
	dataApi.Priority = uint8(dataTf.Priority.ValueInt64())

	dataApi.Enabled = dataTf.Enabled.ValueBool()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Method = mapValueToKey(loadbalancerGroupMethodMap, dataTf.Method.ValueString())
	dataApi.Latitude = dataTf.Latitude.ValueFloat64()
	dataApi.Longitude = dataTf.Longitude.ValueFloat64()
	dataApi.HealthCheck = api.LoadbalancerGroupHealthcheck{
		Enabled: dataTf.HealthcheckEnabled.ValueBool(),
		Path:    dataTf.HealthcheckPath.ValueString(),
	}

	return dataApi
}

func (r *LoadbalancerGroupResource) convertApiToModel(dataApi api.LoadbalancerGroup) (LoadbalancerGroupResourceModel, diag.Diagnostics) {
	dataTf := LoadbalancerGroupResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.LoadbalancerId = types.Int64Value(dataApi.LoadbalancerId)
	dataTf.Priority = types.Int64Value(int64(dataApi.Priority))

	dataTf.Enabled = types.BoolValue(dataApi.Enabled)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Method = types.StringValue(mapKeyToValue(loadbalancerGroupMethodMap, dataApi.Method))
	dataTf.Latitude = types.Float64Value(dataApi.Latitude)
	dataTf.Longitude = types.Float64Value(dataApi.Longitude)
	dataTf.HealthcheckEnabled = types.BoolValue(dataApi.HealthCheck.Enabled)
	dataTf.HealthcheckPath = types.StringValue(dataApi.HealthCheck.Path)

	return dataTf, nil
}
