// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}
var _ resource.ResourceWithConfigValidators = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	client *api.Client
}

type DatabaseResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Url            types.String `tfsdk:"url"`
	RegionsPrimary types.Set    `tfsdk:"regions_primary"`
	RegionsReplica types.Set    `tfsdk:"regions_replica"`
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages an Database in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the database.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the database.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The connection URL for the database.",
			},
			"regions_primary": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"regions_replica": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DatabaseResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf DatabaseResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateDatabase(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create database", err.Error())
		return
	}

	tflog.Trace(ctx, "created database "+dataApi.Name)
	dataTf, diags = databaseApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetDatabase(ctx, data.Id.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching database", err.Error()))
		return
	}

	dataTf, diags := databaseApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateDatabase(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating database", err.Error()))
		return
	}

	dataTf, diags := databaseApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDatabase(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting database", err.Error()))
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	dataApi, err := r.client.GetDatabase(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching database list", err.Error()))
		return
	}

	dataTf, diags := databaseApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DatabaseResource) convertModelToApi(ctx context.Context, dataTf DatabaseResourceModel) api.Database {
	dataApi := api.Database{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Url = dataTf.Url.ValueString()

	{
		elList := dataTf.RegionsPrimary.Elements()
		values := make([]string, 0, len(elList))
		for _, region := range elList {
			values = append(values, region.(types.String).ValueString())
		}
		dataApi.PrimaryRegions = values
	}

	{
		elList := dataTf.RegionsReplica.Elements()
		values := make([]string, 0, len(elList))
		for _, region := range elList {
			values = append(values, region.(types.String).ValueString())
		}
		dataApi.ReplicasRegions = values
	}

	return dataApi
}

func databaseApiToTf(dataApi api.Database) (DatabaseResourceModel, diag.Diagnostics) {
	dataTf := DatabaseResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Url = types.StringValue(dataApi.Url)

	{
		primaryRegions := make([]attr.Value, 0, len(dataApi.PrimaryRegions))
		for _, region := range dataApi.PrimaryRegions {
			primaryRegions = append(primaryRegions, types.StringValue(region))
		}

		set, diags := types.SetValue(types.StringType, primaryRegions)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.RegionsPrimary = set
	}

	{
		replicaRegions := make([]attr.Value, 0, len(dataApi.ReplicasRegions))
		for _, region := range dataApi.ReplicasRegions {
			replicaRegions = append(replicaRegions, types.StringValue(region))
		}

		set, diags := types.SetValue(types.StringType, replicaRegions)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.RegionsReplica = set
	}

	return dataTf, nil
}
