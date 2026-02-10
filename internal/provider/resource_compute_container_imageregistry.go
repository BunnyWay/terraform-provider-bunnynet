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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
)

var _ resource.Resource = &ComputeContainerImageregistryResource{}
var _ resource.ResourceWithImportState = &ComputeContainerImageregistryResource{}

func NewComputeContainerImageregistryResource() resource.Resource {
	return &ComputeContainerImageregistryResource{}
}

type ComputeContainerImageregistryResource struct {
	client *api.Client
}

type ComputeContainerImageregistryResourceModel struct {
	Id       types.Int64  `tfsdk:"id"`
	Registry types.String `tfsdk:"registry"`
	Username types.String `tfsdk:"username"`
	Token    types.String `tfsdk:"token"`
}

func (r *ComputeContainerImageregistryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_container_imageregistry"
}

var computeContainerImageRegistryRegistryOptions = []string{"GitHub", "DockerHub"}

func (r *ComputeContainerImageregistryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages an Image Registry connection for Magic Containers in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the image registry.",
			},
			"registry": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(computeContainerImageRegistryRegistryOptions...),
				},
				Description: generateMarkdownSliceOptions(computeContainerImageRegistryRegistryOptions),
			},
			"username": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The username used to authenticate to the registry.",
			},
			"token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The token used to authenticate to the registry. If you are importing a resource, declare the token as an empty string.",
			},
		},
	}
}

func (r *ComputeContainerImageregistryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ComputeContainerImageregistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf ComputeContainerImageregistryResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	token := dataApi.Token

	dataApi, err := r.client.CreateComputeContainerImageregistry(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create container image registry", err.Error())
		return
	}

	dataApi.Token = token

	tflog.Trace(ctx, "created container image registry "+dataApi.DisplayName+"/"+dataApi.UserName)
	dataTf, diags = r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerImageregistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeContainerImageregistryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetComputeContainerImageregistry(ctx, data.Id.ValueInt64())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching container image registry", err.Error()))
		return
	}

	dataApi.Token = data.Token.ValueString()

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerImageregistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ComputeContainerImageregistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateComputeContainerImageregistry(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating container image registry", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerImageregistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComputeContainerImageregistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeContainerImageregistry(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting container image registry", err.Error()))
	}
}

func (r *ComputeContainerImageregistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error converting ID to integer", err.Error()))
		return
	}

	dataApi, err := r.client.GetComputeContainerImageregistry(ctx, id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching container image registry", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerImageregistryResource) convertModelToApi(ctx context.Context, dataTf ComputeContainerImageregistryResourceModel) api.ComputeContainerImageregistry {
	dataApi := api.ComputeContainerImageregistry{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.DisplayName = dataTf.Registry.ValueString()
	dataApi.UserName = dataTf.Username.ValueString()
	dataApi.Token = dataTf.Token.ValueString()

	return dataApi
}

func (r *ComputeContainerImageregistryResource) convertApiToModel(dataApi api.ComputeContainerImageregistry) (ComputeContainerImageregistryResourceModel, diag.Diagnostics) {
	dataTf := ComputeContainerImageregistryResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Registry = types.StringValue(dataApi.DisplayName)
	dataTf.Username = types.StringValue(dataApi.UserName)
	dataTf.Token = types.StringValue(dataApi.Token)

	return dataTf, nil
}
